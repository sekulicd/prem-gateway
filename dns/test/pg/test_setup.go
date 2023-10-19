package pgtest

import (
	"context"
	"github.com/stretchr/testify/suite"
	pgdb "prem-gateway/dns/internal/infrastructure/storage/pg"
	testutil "prem-gateway/dns/test"
)

var (
	dbSvc *pgdb.Service
	ctx   = context.Background()
)

type PgDbTestSuite struct {
	suite.Suite
}

func (p *PgDbTestSuite) SetupSuite() {
	svc, err := pgdb.NewDBService(pgdb.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "dnsd-db-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/storage/pg/migration",
	})
	if err != nil {
		p.FailNow(err.Error())
	}

	dbSvc = svc

	if err := testutil.SetupDB(); err != nil {
		p.FailNow(err.Error())
	}
}

func (p *PgDbTestSuite) TearDownSuite() {
	if err := testutil.TruncateDB(); err != nil {
		p.FailNow(err.Error())
	}

	dbSvc.Close()
}

func (p *PgDbTestSuite) BeforeTest(suiteName, testName string) {
	if err := testutil.TruncateDB(); err != nil {
		p.FailNow(err.Error())
	}
}

func (p *PgDbTestSuite) AfterTest(suiteName, testName string) {
}
