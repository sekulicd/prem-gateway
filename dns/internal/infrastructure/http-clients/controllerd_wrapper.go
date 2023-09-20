package httpclients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"prem-gateway/dns/internal/core/port"
	"time"
)

type controllerdWrapper struct {
	controllerDaemonUrl string
}

func NewControllerdWrapper(controllerDaemonUrl string) port.ControllerdWrapper {
	return &controllerdWrapper{
		controllerDaemonUrl: controllerDaemonUrl,
	}
}

func (c *controllerdWrapper) DomainProvisioned(
	ctx context.Context, email, domainName string,
) error {
	url := fmt.Sprintf(
		"%s/domain-provisioned?domain=%s&email=%s",
		c.controllerDaemonUrl,
		domainName,
		email,
	)
	return c.sendReq(ctx, url, http.MethodPost)
}

func (c *controllerdWrapper) DomainDeleted(
	ctx context.Context, domainName string,
) error {
	//TODO uncomment once implemented

	//url := fmt.Sprintf(
	//	"%s/domain-deleted?domain=%s",
	//	c.controllerDaemonUrl,
	//	domainName,
	//)
	//return c.sendReq(ctx, url, http.MethodPost)

	return nil
}

func (c *controllerdWrapper) sendReq(ctx context.Context, url string, method string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		nil,
	)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("controllerd returned status code: %v, and error reading body: %v", resp.StatusCode, readErr)
		}

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				err = fmt.Errorf("error closing response body: %v, previous error: %v", closeErr, err)
			}
		}()

		return fmt.Errorf("controllerd returned status code: %v, response: %s", resp.StatusCode, body)
	}

	return nil
}
