package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	letsEncryptStaging = "https://acme-staging-v02.api.letsencrypt.org/directory"
	letsEncryptProd    = "https://acme-v02.api.letsencrypt.org/directory"

	premappService = "premapp"
	premdService   = "premd"
)

var (
	letEncryptProd bool
)

type DnsInfo struct {
	Domain    string `json:"domain"`
	SubDomain string `json:"sub_domain"`
	NodeName  string `json:"node_name"`
	Email     string `json:"email"`
}

func main() {
	serviceNames := os.Getenv("SERVICES")
	services := make([]string, 0)
	if serviceNames != "" {
		services = append(services, strings.Split(serviceNames, ",")...)
	}
	services = append(services, "dnsd")

	letsEncrypt := os.Getenv("LETSENCRYPT_PROD")
	if letsEncrypt != "" {
		letEncryptProd = true
	}

	http.HandleFunc("/domain-provisioned", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			email := r.URL.Query().Get("email")
			domain := r.URL.Query().Get("domain")

			premServices := getPremServicesForRestart(services)
			if len(premServices) > 0 {
				if err := restartServicesWithTls(domain, nil, premServices); err != nil {
					log.Error("Error restarting containers from domainProvisioned : ", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if err := restartServicesWithTls(domain, services, nil); err != nil {
				log.Error("Error restarting containers from domainProvisioned : ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//TODO maybe add health check to all restarted services since services
			//needs to be restarted before traefik can pick up the new labels
			time.Sleep(time.Second * 3)

			if err := restartTraefikWithTls(email); err != nil {
				log.Error("Error restarting traefik from domain-provisioned : ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}()

		if _, err := io.WriteString(w, "OK"); err != nil {
			return
		}
	})

	// TODO expose domainUpdated/Deleted endpoints to restart services without TLS

	log.Info("Starting controller daemon on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Errorf("Controller daemon failed to start: %v", err)
	}
}

func getPremServicesForRestart(srvcs []string) map[string]int {
	svcs := make(map[string]int)
	if contains(srvcs, premdService) {
		resp, err := http.Get("http://premd:8000/v1/services/")
		if err != nil {
			return nil
		}

		if resp == nil {
			return nil
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil
		}

		var premServices []PremService
		if err := json.NewDecoder(resp.Body).Decode(&premServices); err != nil {
			return nil
		}

		for _, v := range premServices {
			if v.Running {
				svcs[v.Id] = v.DefaultPort
			}
		}
	}

	return svcs
}

func contains(slice []string, str string) bool {
	for _, a := range slice {
		if a == str {
			return true
		}
	}
	return false
}

func restartContainer(
	ctx context.Context,
	cli *client.Client,
	containerName string,
	labels map[string]string,
	cmds strslice.StrSlice,
) error {
	containerJson, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return err
	}
	newConfig := containerJson.Config
	//TODO check duplicate labels and cmds
	if len(labels) > 0 {
		newLabels := make(map[string]string)
		for k, v := range newConfig.Labels {
			if !strings.Contains(k, "traefik") {
				newLabels[k] = v
			}
		}
		for k, v := range labels {
			newLabels[k] = v
		}
		newConfig.Labels = newLabels
	}
	if len(cmds) > 0 {
		newConfig.Cmd = append(newConfig.Cmd, cmds...)
	}

	noWaitTimeout := 0
	if err := cli.ContainerStop(
		ctx, containerName, container.StopOptions{Timeout: &noWaitTimeout},
	); err != nil {
		return err
	}
	if err := cli.ContainerRemove(
		ctx, containerName, types.ContainerRemoveOptions{},
	); err != nil {
		//TODO this is workaround for restarting prem services, container removal
		//fails because prem-services are started with --rm flag in prem-daemon
		//and we can't remove them, so we just ignore the error
		//TODO maybe we should check if the container is prem-service and if it is
		//then we should just restart it without removing it
		//sleeping for 5 seconds to wait until container is removed
		log.Warning("Error removing container: ", err)
		time.Sleep(time.Second * 5)
		//return err
	}

	if _, err := cli.ContainerCreate(
		ctx,
		newConfig,
		containerJson.HostConfig,
		&network.NetworkingConfig{
			EndpointsConfig: containerJson.NetworkSettings.Networks,
		},
		nil,
		containerName,
	); err != nil {
		log.Error("Error creating container: ", err)
		return err
	}

	if err := cli.ContainerStart(
		ctx, containerName, types.ContainerStartOptions{},
	); err != nil {
		log.Error("Error starting container: ", err)
		return err
	}

	return nil
}

func restartServicesWithTls(domain string, services []string, premServices map[string]int) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	for _, v := range services {
		switch v {
		case premappService:
			labels := map[string]string{
				"traefik.enable":                                               "true",
				"traefik.http.routers.premapp-http.rule":                       fmt.Sprintf("PathPrefix(`/`) && Host(`%s`)", domain),
				"traefik.http.routers.premapp-http.entrypoints":                "web",
				"traefik.http.routers.premapp-https.rule":                      fmt.Sprintf("PathPrefix(`/`) && Host(`%s`)", domain),
				"traefik.http.routers.premapp-https.entrypoints":               "websecure",
				"traefik.http.routers.premapp-https.tls.certresolver":          "myresolver",
				"traefik.http.middlewares.http-to-https.redirectscheme.scheme": "https",
				"traefik.http.routers.premapp-http.middlewares":                "http-to-https",
				"traefik.http.services.premapp.loadbalancer.server.port":       "8080",
			}

			if err := restartContainer(ctx, cli, v, labels, nil); err != nil {
				return fmt.Errorf("failed to restart container %s: %v", v, err)
			}
		default:
			labels := map[string]string{
				"traefik.enable": "true",
				fmt.Sprintf("traefik.http.routers.%s.rule", v):             fmt.Sprintf("PathPrefix(`/`) && Host(`%s.%s`)", v, domain),
				fmt.Sprintf("traefik.http.routers.%s.entrypoints", v):      "websecure",
				fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", v): "myresolver",
			}

			if err := restartContainer(ctx, cli, v, labels, nil); err != nil {
				return fmt.Errorf("failed to restart container %s: %v", v, err)
			}
		}

		log.Infof("Restarted container %s\n", v)
	}

	for k, v := range premServices {
		labels := map[string]string{
			"traefik.enable": "true",
			fmt.Sprintf("traefik.http.routers.%s-http.rule", k):                    fmt.Sprintf("Host(`%s.%s`)", k, domain),
			fmt.Sprintf("traefik.http.routers.%s-http.entrypoints", k):             "web",
			fmt.Sprintf("traefik.http.routers.%s-https.rule", k):                   fmt.Sprintf("Host(`%s.%s`)", k, domain),
			fmt.Sprintf("traefik.http.routers.%s-https.entrypoints", k):            "websecure",
			fmt.Sprintf("traefik.http.routers.%s-%s.tls.certresolver", k, "https"): "myresolver",
			"traefik.http.middlewares.http-to-https.redirectscheme.scheme":         "https",
			fmt.Sprintf("traefik.http.routers.%s-http.middlewares", k):             "http-to-https",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", k):    strconv.Itoa(v),
		}

		if err := restartContainer(ctx, cli, k, labels, nil); err != nil {
			return fmt.Errorf("failed to restart container %s: %v", k, err)
		}

		log.Infof("Restarted container %s\n", k)
	}

	return nil
}

func restartTraefikWithTls(email string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	traefikLetsEncryptUrl := letsEncryptProd
	if !letEncryptProd {
		traefikLetsEncryptUrl = letsEncryptStaging
	}

	cmds := strslice.StrSlice{
		"--providers.docker=true",
		"--providers.docker.exposedbydefault=false",
		"--accesslog=true",
		"--ping",
		"--entrypoints.web.address=:80",
		"--certificatesresolvers.myresolver.acme.email=" + email,
		"--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json",
		"--certificatesresolvers.myresolver.acme.tlschallenge=true",
		"--certificatesresolvers.myresolver.acme.caserver=" + traefikLetsEncryptUrl,
		"--entrypoints.websecure.address=:443",
	}

	if err := restartContainer(ctx, cli, "traefik", nil, cmds); err != nil {
		return fmt.Errorf("failed to restart container traefik: %v", err)
	}

	log.Info("Restarted container traefik")

	return nil
}

type PremService struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Beta          bool   `json:"beta"`
	Description   string `json:"description"`
	Documentation string `json:"documentation"`
	Icon          string `json:"icon"`
	ModelInfo     struct {
		MemoryRequirements int `json:"memoryRequirements"`
		TokensPerSecond    int `json:"tokensPerSecond"`
	} `json:"modelInfo"`
	Interfaces   []string `json:"interfaces"`
	DockerImages struct {
		Gpu struct {
			Size  int64  `json:"size"`
			Image string `json:"image"`
		} `json:"gpu"`
	} `json:"dockerImages"`
	DefaultPort         int    `json:"defaultPort"`
	DefaultExternalPort int    `json:"defaultExternalPort"`
	RunningPort         int    `json:"runningPort"`
	Banner              string `json:"banner"`
	Running             bool   `json:"running"`
	Downloaded          bool   `json:"downloaded"`
	EnoughMemory        bool   `json:"enoughMemory"`
	EnoughSystemMemory  bool   `json:"enoughSystemMemory"`
	EnoughStorage       bool   `json:"enoughStorage"`
	Command             string `json:"command"`
	DockerImage         string `json:"dockerImage"`
	DockerImageSize     int    `json:"dockerImageSize"`
	Supported           bool   `json:"supported"`
	InvokeMethod        struct {
		Header string `json:"header"`
		SendTo string `json:"baseUrl"`
	} `json:"invokeMethod"`
}
