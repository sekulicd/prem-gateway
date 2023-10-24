package application

import "errors"

var (
	ErrRootKeyExists     = errors.New("root key already exists")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUnauthorizedPath  = errors.New("unauthorized path")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
