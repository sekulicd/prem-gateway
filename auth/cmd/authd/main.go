package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"prem-gateway/auth/internal/config"
	pgdb "prem-gateway/auth/internal/infrastructure/storage/pg"
	httpauthd "prem-gateway/auth/internal/interface/http"
	"syscall"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	svc, err := pgdb.NewRepoService(pgdb.DbConfig{
		DbUser:             config.GetString(config.DbUserKey),
		DbPassword:         config.GetString(config.DbPassKey),
		DbHost:             config.GetString(config.DbHostKey),
		DbPort:             config.GetInt(config.DbPortKey),
		DbName:             config.GetString(config.DbNameKey),
		MigrationSourceURL: config.GetString(config.DbMigrationPathKey),
	})
	if err != nil {
		log.Fatalf("failed to create pgdb service: %s", err)
	}

	authd, err := httpauthd.NewServer(
		config.GetServerAddress(),
		svc,
		config.GetString(config.AdminUserKey),
		config.GetString(config.AdminPassKey),
		config.GetString(config.RootApiKey),
	)
	if err != nil {
		log.Errorf("failed to create prem-gateway auth daemon: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	errC := authd.Start(ctx, stop)
	if err := <-errC; err != nil {
		log.Panicf("prem-gateway auth daemon noticed error while running: %s", err)
	}
}
