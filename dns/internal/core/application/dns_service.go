package application

import (
	"context"
	"errors"
	"prem-gateway/dns/internal/core/domain"
	"prem-gateway/dns/internal/core/port"
	"strings"
)

type DnsService interface {
	CreateDomain(ctx context.Context, dnsInfo DnsInfo) error
	DeleteDomain(ctx context.Context, domainName string) error
	GetDomain(ctx context.Context, domainName string) (DnsInfo, error)
	GetGatewayIp(ctx context.Context) (string, error)
	CheckDnsRecordStatus(ctx context.Context, domainName string) (bool, error)
	GetExistingDomain(ctx context.Context) (*DnsInfo, error)
}

type dnsService struct {
	repositorySvc      domain.RepositoryService
	ipSvc              port.IpService
	controllerdWrapper port.ControllerdWrapper
}

func NewDnsService(
	repositorySvc domain.RepositoryService,
	ipSvc port.IpService,
	controllerdWrapper port.ControllerdWrapper,
) (DnsService, error) {
	return &dnsService{
		repositorySvc:      repositorySvc,
		ipSvc:              ipSvc,
		controllerdWrapper: controllerdWrapper,
	}, nil
}

func (d *dnsService) CreateDomain(ctx context.Context, dnsInfo DnsInfo) error {
	//assumption is that there should be only one domain
	dnsDomain, _ := d.repositorySvc.DnsRepository().Get(ctx, dnsInfo.Domain)
	if dnsDomain != nil {
		return domain.ErrAlreadyExists
	}

	ip, err := d.GetGatewayIp(ctx)
	if err != nil {
		return err
	}
	dnsInfo.Ip = ip

	valid, err := d.ipSvc.VerifyDnsRecord(ctx, dnsInfo.Ip, dnsInfo.Domain)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("dns record not found, check if A record is set correctly")
	}

	if err := d.repositorySvc.DnsRepository().Create(
		ctx, FromAppDnsInfoToDomainDnsInfo(dnsInfo),
	); err != nil {
		return err
	}

	//on initial docker-compose up(main one in proj root) services are
	//started without tls and real subdomains, this will invoke contoller daemon
	//to restart treafik and services with tls/subdomains set
	if err := d.controllerdWrapper.DomainProvisioned(
		ctx, dnsInfo.Email, dnsInfo.Domain,
	); err != nil {
		return err
	}

	return nil
}

func (d *dnsService) DeleteDomain(ctx context.Context, domainName string) error {
	// TODO invoke controllerd to restart services
	return d.repositorySvc.DnsRepository().Delete(ctx, domainName)
}

func (d *dnsService) GetDomain(ctx context.Context, domainName string) (DnsInfo, error) {
	dnsInfo, err := d.repositorySvc.DnsRepository().Get(ctx, domainName)
	if err != nil {
		return DnsInfo{}, err
	}

	return FromDomainDnsInfoToAppDnsInfo(*dnsInfo), nil
}

func (d *dnsService) GetGatewayIp(ctx context.Context) (string, error) {
	ip, err := d.ipSvc.GetHostIp(ctx)
	if err != nil {
		return "", err
	}

	return strings.Replace(ip, "\n", "", -1), nil
}

func (d *dnsService) CheckDnsRecordStatus(
	ctx context.Context, domainName string,
) (bool, error) {
	dnsInfo, err := d.repositorySvc.DnsRepository().Get(ctx, domainName)
	if err != nil {
		return false, err
	}

	return d.ipSvc.VerifyDnsRecord(ctx, dnsInfo.Ip, dnsInfo.Domain)
}

func (d *dnsService) GetExistingDomain(
	ctx context.Context,
) (*DnsInfo, error) {
	dnsInfo, err := d.repositorySvc.DnsRepository().GetExistingDomain(ctx)
	if err != nil {
		if err == domain.ErrEntityNotFound {
			return nil, nil
		}

		return nil, err
	}

	dns := FromDomainDnsInfoToAppDnsInfo(*dnsInfo)

	return &dns, nil
}
