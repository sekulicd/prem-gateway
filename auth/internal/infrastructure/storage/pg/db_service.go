package pgdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"prem-gateway/auth/internal/core/domain"
	"prem-gateway/auth/internal/infrastructure/storage/pg/sqlc/queries"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	postgresDriver             = "pgx"
	insecureDataSourceTemplate = "postgresql://%s:%s@%s:%d/%s?sslmode=disable"

	uniqueViolation = "23505"
	pgxNoRows       = "no rows in result set"
)

type repoService struct {
	pgxPool *pgxpool.Pool
	querier *queries.Queries

	apiKeyRepository domain.ApiKeyRepository
}

func NewRepoService(dbConfig DbConfig) (domain.RepositoryService, error) {
	dataSource := insecureDataSourceStr(dbConfig)

	pgxPool, err := connect(dataSource)
	if err != nil {
		return nil, err
	}

	if err = migrateDb(dataSource, dbConfig.MigrationSourceURL); err != nil {
		return nil, err
	}

	rm := &repoService{
		pgxPool: pgxPool,
		querier: queries.New(pgxPool),
	}

	apiKeyRepository := NewIdentityRepositoryImpl(rm.querier, rm.execTx)
	rm.apiKeyRepository = apiKeyRepository

	return rm, nil
}

func (s *repoService) ApiKeyRepository() domain.ApiKeyRepository {
	return s.apiKeyRepository
}

func (s *repoService) Close() {
	s.pgxPool.Close()
}

func (s *repoService) execTx(
	ctx context.Context,
	txBody func(*queries.Queries) error,
) error {
	conn, err := s.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op.
	defer func() {
		err := tx.Rollback(ctx)
		switch {
		// If the tx was already closed (it was successfully executed)
		// we do not need to log that error.
		case errors.Is(err, pgx.ErrTxClosed):
			return

		// If this is an unexpected error, log it.
		case err != nil:
			log.Errorf("unable to rollback db tx: %v", err)
		}
	}()

	if err := txBody(s.querier.WithTx(tx)); err != nil {
		return err
	}

	// Commit transaction.
	return tx.Commit(ctx)
}

type DbConfig struct {
	DbUser             string
	DbPassword         string
	DbHost             string
	DbPort             int
	DbName             string
	MigrationSourceURL string
}

func connect(dataSource string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), dataSource)
}

func migrateDb(dataSource, migrationSourceUrl string) error {
	pg := postgres.Postgres{}

	d, err := pg.Open(dataSource)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationSourceUrl,
		postgresDriver,
		d,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func insecureDataSourceStr(dbConfig DbConfig) string {
	return fmt.Sprintf(
		insecureDataSourceTemplate,
		dbConfig.DbUser,
		dbConfig.DbPassword,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)
}
