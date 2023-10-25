package httphandler

import "prem-gateway/auth/internal/core/application"

type CreateApiKey struct {
	Service          string `json:"service_name"`
	RequestsPerRange int    `json:"requests_per_range"`
	RangeInMinutes   int    `json:"range_in_minutes"`
}

func ToAppCreateApiKeyInfo(req CreateApiKey) application.CreateApiKeyReq {
	rangeInSeconds := req.RangeInMinutes * 60
	return application.CreateApiKeyReq{
		Service:          req.Service,
		RequestsPerRange: req.RequestsPerRange,
		RangeInSeconds:   rangeInSeconds,
	}
}
