package httphandler

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"prem-gateway/auth/internal/core/application"
	"strings"
)

type AuthHandler interface {
	LogIn(c *gin.Context)
	CreateApiKey(c *gin.Context)
	GetServiceApiKey(c *gin.Context)
	IsRequestAllowed(c *gin.Context)
}

type authHandler struct {
	apiKeySvc application.ApiKeyService
	authSvc   application.AuthService
}

func NewAuthHandler(
	apiKeySvc application.ApiKeyService,
	authSvc application.AuthService,
) (AuthHandler, error) {
	return &authHandler{
		apiKeySvc: apiKeySvc,
		authSvc:   authSvc,
	}, nil
}

func (a *authHandler) LogIn(c *gin.Context) {
	user := c.Query("user")
	pass := c.Query("pass")

	apiKey, err := a.authSvc.AuthAdmin(c, user, pass)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
	})
}

func (a *authHandler) CreateApiKey(c *gin.Context) {
	var req CreateApiKey
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := a.apiKeySvc.CreateApiKey(c, ToAppCreateApiKeyInfo(req))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": id,
	})
}

func (a *authHandler) GetServiceApiKey(c *gin.Context) {
	service := c.Param("service")

	apiKey, err := a.apiKeySvc.GetServiceApiKey(c, service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
	})
}

func (a *authHandler) IsRequestAllowed(c *gin.Context) {
	apiKey := c.GetHeader("Authorization")
	uri := c.GetHeader("X-Forwarded-Uri")
	forwardedFor := c.GetHeader("X-Forwarded-For")

	log.Infof("Authorization header: %s", apiKey)
	log.Infof("X-Forwarded-Uri header: %s", uri)
	log.Infof("X-Forwarded-For header: %s", forwardedFor)

	service := extractService(forwardedFor, uri)
	if service == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrServiceNotFound,
		})
		return
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrApiKeyNotProvided,
		})
		return
	}

	if err := a.apiKeySvc.AllowRequest(apiKey, service); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func extractService(host string, uri string) string {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return "" // Could not parse URI
	}

	path := parsedUri.Path
	if net.ParseIP(host) == nil {
		parts := strings.Split(host, ".")
		if len(parts) > 1 {
			return parts[0]
		}
	} else {
		uriParts := strings.Split(path, "/")
		if len(uriParts) > 1 {
			return uriParts[1]
		}
	}

	return ""
}
