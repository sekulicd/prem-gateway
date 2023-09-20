package pgdb

import (
	"context"
	"database/sql"
	"github.com/jackc/pgconn"
	"prem-gateway/dns/internal/core/domain"
	"prem-gateway/dns/internal/infrastructure/storage/pg/sqlc/queries"
)

type dnsRepositoryImpl struct {
	querier *queries.Queries
}

func NewDnsRepositoryImpl(querier *queries.Queries) domain.DnsRepository {
	return &dnsRepositoryImpl{
		querier: querier,
	}
}

func (d *dnsRepositoryImpl) Create(
	ctx context.Context, dnsInfo domain.DnsInfo,
) error {
	var subDomain, ip, nodeName, email sql.NullString
	if dnsInfo.SubDomain != "" {
		subDomain = sql.NullString{
			String: dnsInfo.SubDomain,
			Valid:  true,
		}
	}
	if dnsInfo.Ip != "" {
		ip = sql.NullString{
			String: dnsInfo.Ip,
			Valid:  true,
		}
	}
	if dnsInfo.NodeName != "" {
		nodeName = sql.NullString{
			String: dnsInfo.NodeName,
			Valid:  true,
		}
	}
	if dnsInfo.Email != "" {
		email = sql.NullString{
			String: dnsInfo.Email,
			Valid:  true,
		}
	}

	if err := d.querier.InsertDnsInfo(ctx, queries.InsertDnsInfoParams{
		Domain:    dnsInfo.Domain,
		SubDomain: subDomain,
		Ip:        ip,
		NodeName:  nodeName,
		Email:     email,
	}); err != nil {
		if pqErr := err.(*pgconn.PgError); pqErr != nil {
			if pqErr.Code == uniqueViolation {
				return nil
			} else {
				return err
			}
		}
	}

	return nil
}

func (d *dnsRepositoryImpl) Delete(
	ctx context.Context, domain string,
) error {
	return d.querier.DeleteDnsInfo(ctx, domain)
}

func (d *dnsRepositoryImpl) Get(
	ctx context.Context,
	domainName string,
) (*domain.DnsInfo, error) {
	dnsInfo, err := d.querier.GetDnsInfo(ctx, domainName)
	if err != nil {
		if err != nil {
			if err.Error() == pgxNoRows {
				return nil, domain.ErrEntityNotFound
			}

			return nil, err
		}
	}

	var subDomain, ip, nodeName, email string
	if dnsInfo.SubDomain.Valid {
		subDomain = dnsInfo.SubDomain.String
	}
	if dnsInfo.Ip.Valid {
		ip = dnsInfo.Ip.String
	}
	if dnsInfo.NodeName.Valid {
		nodeName = dnsInfo.NodeName.String
	}
	if dnsInfo.Email.Valid {
		email = dnsInfo.Email.String
	}

	return &domain.DnsInfo{
		Domain:    dnsInfo.Domain,
		SubDomain: subDomain,
		Ip:        ip,
		NodeName:  nodeName,
		Email:     email,
	}, nil
}

func (d *dnsRepositoryImpl) GetExistingDomain(ctx context.Context) (*domain.DnsInfo, error) {
	dns, err := d.querier.GetExistDnsInfo(ctx)
	if err != nil {
		if err.Error() == pgxNoRows {
			return nil, domain.ErrEntityNotFound
		}

		return nil, err
	}

	var subDomain, ip, nodeName, email string
	if dns.SubDomain.Valid {
		subDomain = dns.SubDomain.String
	}
	if dns.Ip.Valid {
		ip = dns.Ip.String
	}
	if dns.NodeName.Valid {
		nodeName = dns.NodeName.String
	}
	if dns.Email.Valid {
		email = dns.Email.String
	}

	return &domain.DnsInfo{
		Domain:    dns.Domain,
		SubDomain: subDomain,
		Ip:        ip,
		NodeName:  nodeName,
		Email:     email,
	}, nil
}
