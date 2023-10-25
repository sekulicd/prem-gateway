package httpdnsd

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"prem-gateway/dns/internal/core/application"
	"prem-gateway/dns/internal/core/domain"
	httphandler "prem-gateway/dns/internal/interface/http/handler"
	"time"
)

const (
	shutdownTimeout = time.Second * 5
)

type Server interface {
	Start(ctx context.Context, stop context.CancelFunc) <-chan error
	Stop() error
	Router() http.Handler
}

type server struct {
	serverAddress string
	opts          serverOptions
	dnsHandler    httphandler.DNSHandler
	dnsSvc        application.DnsService
}

func NewServer(
	serverAddress string,
	repositorySvc domain.RepositoryService,
	controllerDaemonUrl string,
	opts ...ServerOption,
) (Server, error) {
	options := defaultServerOptions(controllerDaemonUrl)
	for _, o := range opts {
		if err := o.apply(&options); err != nil {
			return nil, err
		}
	}

	dnsSvc, err := application.NewDnsService(
		repositorySvc, options.ipSvc, options.controllerdWrapper,
	)
	if err != nil {
		return nil, err
	}

	dnsHandler, err := httphandler.NewDNSHandler(dnsSvc)
	if err != nil {
		return nil, err
	}

	return &server{
		serverAddress: serverAddress,
		opts:          options,
		dnsHandler:    dnsHandler,
		dnsSvc:        dnsSvc,
	}, nil
}

func (s *server) Start(ctx context.Context, stop context.CancelFunc) <-chan error {
	errCh := make(chan error)

	httpServer := &http.Server{
		Addr:    s.serverAddress,
		Handler: s.Router(),
	}

	go func() {
		<-ctx.Done()

		log.Info("shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		defer func() {
			stop()
			cancel()
			close(errCh)
		}()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctxTimeout); err != nil {
			errCh <- err
		}

		log.Info("prem-gateway dns daemon graceful shutdown completed")
	}()

	go func() {
		log.Infof("prem-gateway dns daemon listening and serving at: %v", s.serverAddress)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	return errCh
}

func (s *server) Stop() error {
	return nil
}

func (s *server) Router() http.Handler {
	ginEngine := gin.Default()
	ginEngine.Use(gin.Recovery())
	ginEngine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:1420") // Replace with your frontend origin
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	ginEngine.POST("/dns", s.dnsHandler.CreateDnsInfo)
	ginEngine.DELETE("/dns/:domain", s.dnsHandler.DeleteDnsInfo)
	ginEngine.GET("/dns/:domain", s.dnsHandler.GetDnsInfo)
	ginEngine.GET("/dns/status/:domain", s.dnsHandler.CheckDnsStatus)
	ginEngine.GET("/dns/ip", s.dnsHandler.GetGatewayIp)
	ginEngine.GET("/dns/check", s.dnsHandler.Check)
	ginEngine.GET("/dns/existing", s.dnsHandler.GetExistingDns)
	ginEngine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return ginEngine
}
