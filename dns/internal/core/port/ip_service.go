package port

import "context"

type IpService interface {
	VerifyDnsRecord(ctx context.Context, ip, domainName string) (bool, error)
	GetHostIp(ctx context.Context) (string, error)
}
