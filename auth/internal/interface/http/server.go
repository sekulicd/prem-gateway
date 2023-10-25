package httpauthd

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"prem-gateway/auth/internal/core/application"
	"prem-gateway/auth/internal/core/domain"
	httphandler "prem-gateway/auth/internal/interface/http/handler"
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
	authHandler   httphandler.AuthHandler
}

func NewServer(
	serverAddress string,
	repositorySvc domain.RepositoryService,
	adminUser string,
	adminPass string,
	rootKeyApiKey string,
	opts ...ServerOption,
) (Server, error) {
	options := defaultServerOptions()
	for _, o := range opts {
		if err := o.apply(&options); err != nil {
			return nil, err
		}
	}

	apiKeySvc, err := application.NewApiKeyService(
		context.Background(), rootKeyApiKey, repositorySvc,
	)
	if err != nil {
		return nil, err
	}

	authSvc := application.NewAuthService(adminUser, adminPass, repositorySvc)

	authHandler, err := httphandler.NewAuthHandler(apiKeySvc, authSvc)
	if err != nil {
		return nil, err
	}

	return &server{
		serverAddress: serverAddress,
		opts:          options,
		authHandler:   authHandler,
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

		log.Info("prem-gateway auth daemon graceful shutdown completed")
	}()

	go func() {
		log.Infof("prem-gateway auth daemon listening and serving at: %v", s.serverAddress)

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
	ginEngine := gin.New()
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

	ginEngine.GET("/auth/login", s.authHandler.LogIn)
	ginEngine.GET("/auth/verify", s.authHandler.IsRequestAllowed)
	ginEngine.POST("/auth/api-key", s.authHandler.CreateApiKey)
	ginEngine.GET("/auth/api-key/service", s.authHandler.GetServiceApiKey)
	return ginEngine
}
