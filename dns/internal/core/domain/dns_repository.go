package domain

import "context"

type DnsRepository interface {
	Create(ctx context.Context, dnsInfo DnsInfo) error
	Delete(ctx context.Context, domainName string) error
	Get(ctx context.Context, domainName string) (*DnsInfo, error)
	GetExistingDomain(ctx context.Context) (*DnsInfo, error)
}
