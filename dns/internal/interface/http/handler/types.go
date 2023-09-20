package httphandler

import "prem-gateway/dns/internal/core/application"

type DnsInfo struct {
	Domain   string `json:"domain"`
	Ip       string `json:"ip"`
	NodeName string `json:"node_name"`
	Email    string `json:"email"`
}

func FromHandlerDnsInfoToAppDnsInfo(hdi DnsInfo) application.DnsInfo {
	return application.DnsInfo{
		Domain:   hdi.Domain,
		Ip:       hdi.Ip,
		NodeName: hdi.NodeName,
		Email:    hdi.Email,
	}
}

func FromAppDnsInfoToHandlerDnsInfo(adi application.DnsInfo) DnsInfo {
	return DnsInfo{
		Domain:   adi.Domain,
		Ip:       adi.Ip,
		NodeName: adi.NodeName,
		Email:    adi.Email,
	}
}

type SuccessResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
