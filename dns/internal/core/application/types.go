package application

import (
	"fmt"
	"prem-gateway/dns/internal/core/domain"
)

type DnsInfo struct {
	Domain   string
	Ip       string
	NodeName string
	Email    string
}

func FromAppDnsInfoToDomainDnsInfo(dnsInfo DnsInfo) domain.DnsInfo {
	return domain.DnsInfo{
		Domain:    dnsInfo.Domain,
		SubDomain: fmt.Sprintf("*.%s", dnsInfo.Domain),
		Ip:        dnsInfo.Ip,
		NodeName:  dnsInfo.NodeName,
		Email:     dnsInfo.Email,
	}
}

func FromDomainDnsInfoToAppDnsInfo(dnsInfo domain.DnsInfo) DnsInfo {
	return DnsInfo{
		Domain:   dnsInfo.Domain,
		Ip:       dnsInfo.Ip,
		NodeName: dnsInfo.NodeName,
		Email:    dnsInfo.Email,
	}
}
