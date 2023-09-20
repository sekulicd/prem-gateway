package httpclients

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"prem-gateway/dns/internal/core/port"
	"time"
)

type ipService struct {
}

func NewIpService() port.IpService {
	return &ipService{}
}

func (i *ipService) VerifyDnsRecord(
	ctx context.Context, expectedIP, domainName string,
) (bool, error) {
	ips, err := net.LookupIP(domainName)
	if err != nil {
		return false, fmt.Errorf("DNS record not found for domain: %v", domainName)
	} else {
		for _, ip := range ips {
			if ip.String() == expectedIP {
				return true, nil
			}
		}
		return false, errors.New("DNS record found, but IP does not match")
	}
}

func (i *ipService) GetHostIp(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://ifconfig.io", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
