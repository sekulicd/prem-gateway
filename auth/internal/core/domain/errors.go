package domain

import "errors"

var (
	ErrApiKeyExistForService = errors.New("api key already exists for service")
	ErrApiKeyNotFound        = errors.New("api key not found")
)
