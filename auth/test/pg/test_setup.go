package pgtest

import (
	"context"
	"prem-gateway/auth/internal/core/domain"
	pgdb "prem-gateway/auth/internal/infrastructure/storage/pg"
	testutil "prem-gateway/auth/test"

	"github.com/stretchr/testify/suite"
)

var (
	repositorySvc domain.RepositoryService
	ctx           = context.Background()
)

type PgDbTestSuite struct {
	suite.Suite
}

func (p *PgDbTestSuite) SetupSuite() {
	rsvc, err := pgdb.NewRepoService(pgdb.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "authd-db-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/storage/pg/migration",
	})
	if err != nil {
		p.FailNow(err.Error())
	}

	repositorySvc = rsvc

	if err := testutil.SetupDB(); err != nil {
		p.FailNow(err.Error())
	}
}

func (p *PgDbTestSuite) TearDownSuite() {
	if err := testutil.TruncateDB(); err != nil {
		p.FailNow(err.Error())
	}
}

func (p *PgDbTestSuite) BeforeTest(suiteName, testName string) {
	if err := testutil.TruncateDB(); err != nil {
		p.FailNow(err.Error())
	}
}

func (p *PgDbTestSuite) AfterTest(suiteName, testName string) {
}
