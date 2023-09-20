package httpdnsd

import (
	"prem-gateway/dns/internal/core/port"
	httpclients "prem-gateway/dns/internal/infrastructure/http-clients"
)

type ServerOption interface {
	apply(*serverOptions) error
}

type serverOptions struct {
	ipSvc              port.IpService
	controllerdWrapper port.ControllerdWrapper
}

func defaultServerOptions(controllerDaemonUrl string) serverOptions {
	ipSvc := httpclients.NewIpService()
	controllerdWrapper := httpclients.NewControllerdWrapper(controllerDaemonUrl)
	return serverOptions{
		ipSvc:              ipSvc,
		controllerdWrapper: controllerdWrapper,
	}
}

type funcServerOption struct {
	f func(*serverOptions) error
}

func (fdo *funcServerOption) apply(do *serverOptions) error {
	return fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions) error) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithIpService(ipSvc port.IpService) ServerOption {
	return newFuncServerOption(func(o *serverOptions) error {
		o.ipSvc = ipSvc
		return nil
	})
}

func WithControllerdWrapper(
	controllerdWrapper port.ControllerdWrapper,
) ServerOption {
	return newFuncServerOption(func(o *serverOptions) error {
		o.controllerdWrapper = controllerdWrapper
		return nil
	})
}
