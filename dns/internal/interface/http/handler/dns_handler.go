package httphandler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"prem-gateway/dns/internal/core/application"
	"prem-gateway/dns/internal/core/domain"
)

type DNSHandler interface {
	CreateDnsInfo(c *gin.Context)
	DeleteDnsInfo(c *gin.Context)
	GetDnsInfo(c *gin.Context)
	CheckDnsStatus(c *gin.Context)
	GetGatewayIp(c *gin.Context)
	GetExistingDns(c *gin.Context)
	Check(c *gin.Context)
}

type dnsHandler struct {
	dnsSvc application.DnsService
}

func NewDNSHandler(dnsSvc application.DnsService) (DNSHandler, error) {
	return &dnsHandler{
		dnsSvc: dnsSvc,
	}, nil
}

// CreateDnsInfo godoc
// @Summary Creates a new DNS record
// @Description This endpoint creates a new DNS record based on the provided information
// @Tags dns
// @Accept json
// @Produce json
// @Param DnsCreateReq body DnsCreateReq true "dns information"
//
//	@Success		200		{object}	SuccessResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//
// @Router /dns [post]
func (d *dnsHandler) CreateDnsInfo(c *gin.Context) {
	var dnsCreateReq DnsCreateReq
	if err := c.ShouldBindJSON(&dnsCreateReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := d.dnsSvc.CreateDomain(
		c.Request.Context(),
		FromHandlerDnsInfoToAppDnsInfo(dnsCreateReq),
	); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Status: "success"})
}

// DeleteDnsInfo godoc
// @Summary Deletes a DNS record
// @Description This endpoint deletes a DNS record based on the provided domain name
// @Tags dns
// @Accept json
// @Produce json
// @Param domain path string true "Domain Name"
//
//	@Success		200		{object}	SuccessResponse	"Returns status of operation"
//	@Failure		400		{object}	ErrorResponse	"Returns error message for invalid input"
//	@Failure		500		{object}	ErrorResponse	"Returns error message for server error"
//
// @Router /dns/{domain} [delete]
func (d *dnsHandler) DeleteDnsInfo(c *gin.Context) {
	domainName := c.Param("domain")
	if domainName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "domain is empty"})
		return
	}

	if err := d.dnsSvc.DeleteDomain(
		c.Request.Context(),
		domainName,
	); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Status: "success"})
}

// GetDnsInfo godoc
// @Summary Retrieves a DNS record
// @Description This endpoint retrieves a DNS record based on the provided domain name
// @Tags dns
// @Accept json
// @Produce json
// @Param domain path string true "Domain Name"
//
//	@Success		200		{object}	DnsInfo		"Returns the DNS record"
//	@Failure		400		{object}	ErrorResponse	"Returns error message for invalid input"
//	@Failure		404		{object}	ErrorResponse	"Returns error message for record not found"
//	@Failure		500		{object}	ErrorResponse	"Returns error message for server error"
//
// @Router /dns/{domain} [get]
func (d *dnsHandler) GetDnsInfo(c *gin.Context) {
	domainName := c.Param("domain")
	if domainName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "domain is empty"})
		return
	}

	dnsInfo, err := d.dnsSvc.GetDomain(c.Request.Context(), domainName)
	if err != nil {
		if err == domain.ErrEntityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, FromAppDnsInfoToHandlerDnsInfo(dnsInfo))
}

// CheckDnsStatus godoc
// @Summary Check status of a DNS record
// @Description This endpoint checks the status of a DNS record based on the provided domain name
// @Tags dns
// @Accept json
// @Produce json
// @Param domain path string true "Domain Name"
//
//	@Success		200		{object}	bool		"Returns true if the DNS record is valid, false otherwise"
//	@Failure		400		{object}	ErrorResponse	"Returns error message for invalid input"
//	@Failure		404		{object}	ErrorResponse	"Returns error message for record not found"
//	@Failure		500		{object}	ErrorResponse	"Returns error message for server error"
//
// @Router /dns/status/{domain} [get]
func (d *dnsHandler) CheckDnsStatus(c *gin.Context) {
	domainName := c.Param("domain")
	if domainName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "domain is empty"})
		return
	}

	valid, err := d.dnsSvc.CheckDnsRecordStatus(c.Request.Context(), domainName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "dns record not found"})
		return
	}

	c.JSON(http.StatusOK, valid)
}

// GetGatewayIp godoc
// @Summary Retrieves the IP address of the Gateway
// @Description This endpoint retrieves the IP address of the Gateway
// @Tags dns
// @Accept json
// @Produce json
//
//	@Success		200		{object}	string		"Returns IP address of the Gateway"
//	@Failure		500		{object}	ErrorResponse	"Returns error message for server error"
//
// @Router /dns/ip [get]
func (d *dnsHandler) GetGatewayIp(c *gin.Context) {
	ipAddr, err := d.dnsSvc.GetGatewayIp(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.String(http.StatusOK, ipAddr)
}

// GetExistingDns godoc
// @Summary Retrieves the existing DNS record
// @Description This endpoint retrieves the existing DNS record
// @Tags dns
// @Accept json
// @Produce json
//
//	@Success		200		{object}	DnsInfo		"Returns the existing DNS record"
//	@Failure		500		{object}	ErrorResponse	"Returns error message for server error"
//
// @Router /dns/existing [get]
func (d *dnsHandler) GetExistingDns(c *gin.Context) {
	dnsInfo, err := d.dnsSvc.GetExistingDomain(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if dnsInfo != nil {
		c.JSON(http.StatusOK, FromAppDnsInfoToHandlerDnsInfo(*dnsInfo))
		return
	}

	c.JSON(http.StatusOK, nil)
}

// Check godoc
// @Summary Check if the service is up and running
// @Description This endpoint checks if the service is up and running
// @Tags dns
// @Accept json
// @Produce json
// @Success 200
// @Router /dns/check [get]
func (d *dnsHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
