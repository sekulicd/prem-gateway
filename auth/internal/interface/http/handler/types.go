package httphandler

import "prem-gateway/auth/internal/core/application"

type CreateApiKey struct {
	Service          string
	RequestsPerRange int
	RangeInMinutes   int
}

func ToAppCreateApiKeyInfo(req CreateApiKey) application.CreateApiKeyReq {
	rangeInSeconds := req.RangeInMinutes * 60
	return application.CreateApiKeyReq{
		Service:          req.Service,
		RequestsPerRange: req.RequestsPerRange,
		RangeInSeconds:   rangeInSeconds,
	}
}
