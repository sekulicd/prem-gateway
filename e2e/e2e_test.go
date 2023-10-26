package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	client = &http.Client{}
)

// TestE2eFeatures is an end-to-end test with purpose of documenting DNS/TLS/Auth
// features and traefik routing.
// First it queries premd service to get all services and their base urls and it
// shows how traefik routes requests to by path.
// Then it creates DNS record for domain and it shows how traefik routes requests
// to by subdomains.
// In the end it shows how to run prem-service and how to make request to it.
// It is assumed that prem-service gpt4all-lora-q4 docker image is downloaded already
// since doing it in test would take too much time.
// It is necessary to relate A records to prem-gateway IP address in DNS provider.
// Eg. considering domain example.com and prem-gateway IP address
// 1. Create A record for example.com with value prem-gateway IP address
// 2. Create A record for *.example.com with value prem-gateway IP address
// Env. variables:
// PREM_GATEWAY_IP - IP address of prem-gateway
// USER_NAME - username for basic auth
// PASSWORD - password for basic auth
// DOMAIN - domain for DNS record
// EMAIL - email for DNS record
func TestE2eFeatures(t *testing.T) {
	premGatewayIP := os.Getenv("PREM_GATEWAY_IP")
	if premGatewayIP == "" {
		t.Fatal("PREM_GATEWAY_IP environment variable not set")
	}

	userName := os.Getenv("USER_NAME")
	if userName == "" {
		t.Fatal("USER_NAME environment variable not set")
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		t.Fatal("PASSWORD environment variable not set")
	}

	//GET ROOT API KEY
	resp, err := http.Get(
		fmt.Sprintf(
			"http://%s/%s?user=%s&pass=%s",
			premGatewayIP,
			"authd/auth/login",
			userName,
			password,
		),
	)
	assert.NoError(t, err)
	apiKey := make(map[string]string)
	err = json.NewDecoder(resp.Body).Decode(&apiKey)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	rootApiKey := apiKey["api_key"]
	assert.NotEmpty(t, rootApiKey)

	servicesUrls := make([]ExtractedFields, 0)

	url := fmt.Sprintf("http://%s/%s", premGatewayIP, "premd/v1/services/")
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&servicesUrls)
	assert.NoError(t, err)

	for _, v := range servicesUrls {
		assert.Equal(t, v.BaseUrl, fmt.Sprintf("http://%s/%s", premGatewayIP, v.ServiceId))
	}

	url = fmt.Sprintf("http://%s/%s", premGatewayIP, "dnsd/dns/existing")
	req, err = http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, string(bodyBytes), "null")

	//CREATE DOMAIN
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		t.Fatal("DOMAIN environment variable not set")
	}

	email := os.Getenv("EMAIL")
	if email == "" {
		t.Fatal("EMAIL environment variable not set")
	}

	dnsCreateReq := DnsCreateReq{
		Domain:   domain,
		Email:    email,
		NodeName: "prem-gateway",
	}

	jsonValue, err := json.Marshal(dnsCreateReq)
	assert.NoError(t, err)

	url = fmt.Sprintf("http://%s/%s", premGatewayIP, "dnsd/dns")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	time.Sleep(10 * time.Second) // Wait for controller to restart services

	// GET PREMD SERVICES VIA SUBDOMAIN
	url = fmt.Sprintf("https://%s.%s/%s", "premd", dnsCreateReq.Domain, "v1/services/")
	req, err = http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	servicesUrls = make([]ExtractedFields, 0)
	err = json.NewDecoder(resp.Body).Decode(&servicesUrls)
	assert.NoError(t, err)

	for _, v := range servicesUrls {
		assert.Equal(t, v.BaseUrl, fmt.Sprintf("https://%s.%s", v.ServiceId, dnsCreateReq.Domain))
	}

	url = fmt.Sprintf("https://%s.%s/%s", "dnsd", dnsCreateReq.Domain, "dns/existing")
	req, err = http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	checkDns := DnsInfo{}
	bodyBytes, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(bodyBytes, &checkDns)
	assert.NoError(t, err)
	assert.Equal(t, checkDns.Domain, dnsCreateReq.Domain)

	// Assume prem-service gpt4all-lora-q4 docker image is downloaded previously

	//RUN SERVICE
	runService := ExtractedFields{
		ServiceId: "gpt4all-lora-q4",
	}
	jsonValue, err = json.Marshal(runService)
	assert.NoError(t, err)

	url = fmt.Sprintf("https://%s.%s/%s", "premd", dnsCreateReq.Domain, "v1/run-service/")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(5 * time.Second) // Wait for prem-service to start

	//make POST on POST https://gpt4all-lora-q4.dusansekulic.me/v1/chat/completions with body
	chatRequest := ChatRequest{
		Model: "llama-2-13b-chat",
		Messages: []Message{
			{
				Role:    "user",
				Content: "hello",
			},
		},
		Stream:           true,
		Temperature:      0.2,
		MaxTokens:        256,
		TopP:             0.95,
		FrequencyPenalty: 0,
		N:                1,
		PresencePenalty:  0,
	}
	jsonValue, err = json.Marshal(chatRequest)
	assert.NoError(t, err)

	resp, err = http.Post(
		fmt.Sprintf("https://%s.%s/%s", "gpt4all-lora-q4", dnsCreateReq.Domain, "v1/chat/completions"),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	createApiKeyReq := CreateApiKey{
		ServiceName:      "gpt4all-lora-q4",
		RequestsPerRange: 10,
		RangeInMinutes:   1,
	}

	jsonValue, err = json.Marshal(createApiKeyReq)
	assert.NoError(t, err)
	url = fmt.Sprintf("https://%s.%s/%s", "authd", dnsCreateReq.Domain, "auth/api-key")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	apiKey = make(map[string]string)
	err = json.NewDecoder(resp.Body).Decode(&apiKey)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	serviceApiKey := apiKey["api_key"]

	url = fmt.Sprintf("https://%s.%s/%s", "gpt4all-lora-q4", dnsCreateReq.Domain, "v1/chat/completions")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Add("Authorization", serviceApiKey)
	resp, err = client.Do(req)
	bodyBytes, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	t.Log(string(bodyBytes))
}

type ExtractedFields struct {
	ServiceId string `json:"id"`
	BaseUrl   string `json:"baseUrl"`
}

type DnsInfo struct {
	Domain   string `json:"domain"`
	Ip       string `json:"ip"`
	NodeName string `json:"nodeName"`
	Email    string `json:"email"`
}

type DnsCreateReq struct {
	Domain   string `json:"domain"`
	Email    string `json:"email"`
	NodeName string `json:"nodeName"`
}

type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream"`
	Temperature      float64   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             float64   `json:"top_p"`
	FrequencyPenalty int       `json:"frequency_penalty"`
	N                int       `json:"n"`
	PresencePenalty  int       `json:"presence_penalty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CreateApiKey struct {
	ServiceName      string `json:"service_name"`
	RequestsPerRange int    `json:"requests_per_range"`
	RangeInMinutes   int    `json:"range_in_minutes"`
}
