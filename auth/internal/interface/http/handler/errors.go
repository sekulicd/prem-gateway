package httphandler

import "errors"

var (
	ErrServiceNotFound   = errors.New("could not extract service from request")
	ErrApiKeyNotProvided = errors.New("API key not provided")
)
