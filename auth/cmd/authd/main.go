package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	//_ "prem-gateway/auth/docs"
	"prem-gateway/auth/internal/config"
	pgdb "prem-gateway/auth/internal/infrastructure/storage/pg"
	httpauthd "prem-gateway/auth/internal/interface/http"
	"syscall"
)

// @title Dns Daemon API
// @description     DNS Daemon is designed to manage Domain Name System (DNS) records. <br />It exposes a RESTful API that allows for the creation, modification, retrieval, and deletion of DNS information, as well as checking the status of a DNS entry. <br /> The DNS information includes attributes such as domain, subdomain, A records, and node names.
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

//TODO add swagger, check dnsd swagger
//add more comments in general to auth
//add http tests
//rename NewDBService in dns, check mock
//update docker
//traefik labels
//controllerd labels
//remove basic auth
//handle panic recovery
