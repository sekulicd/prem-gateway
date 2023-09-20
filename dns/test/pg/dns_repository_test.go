package pgtest

import "prem-gateway/dns/internal/core/domain"

func (p *PgDbTestSuite) TestDnsRepository() {
	dnsInfo, err := dbSvc.DnsRepository().Get(ctx, "dummy")
	p.EqualError(err, domain.ErrEntityNotFound.Error())
	p.Nil(dnsInfo)

	dnsInfo = &domain.DnsInfo{
		Domain:    "example.com",
		SubDomain: "*example.com",
		Ip:        "10.10.10.10",
		NodeName:  "node1",
		Email:     "test@gmail.com",
	}

	err = dbSvc.DnsRepository().Create(ctx, *dnsInfo)
	p.NoError(err)

	dnsInfo, err = dbSvc.DnsRepository().Get(ctx, "example.com")
	p.NoError(err)
	p.Equal("example.com", dnsInfo.Domain)
	p.Equal("*example.com", dnsInfo.SubDomain)
	p.Equal("10.10.10.10", dnsInfo.Ip)
	p.Equal("node1", dnsInfo.NodeName)
	p.Equal("test@gmail.com", dnsInfo.Email)

	dns, err := dbSvc.DnsRepository().GetExistingDomain(ctx)
	p.NoError(err)
	p.Equal("example.com", dns.Domain)
	p.Equal("*example.com", dns.SubDomain)
	p.Equal("10.10.10.10", dnsInfo.Ip)
	p.Equal("node1", dnsInfo.NodeName)
	p.Equal("test@gmail.com", dnsInfo.Email)

	err = dbSvc.DnsRepository().Create(ctx, *dnsInfo)
	p.NoError(err)

	err = dbSvc.DnsRepository().Delete(ctx, "dummy")
	p.NoError(err)

	err = dbSvc.DnsRepository().Delete(ctx, "example.com")
	p.NoError(err)

	dnsInfo, err = dbSvc.DnsRepository().Get(ctx, "example.com")
	p.EqualError(err, domain.ErrEntityNotFound.Error())
	p.Nil(dnsInfo)
}
