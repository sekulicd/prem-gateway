package port

import "context"

type ControllerdWrapper interface {
	DomainProvisioned(ctx context.Context, email, domainName string) error
	DomainDeleted(ctx context.Context, domainName string) error
}
