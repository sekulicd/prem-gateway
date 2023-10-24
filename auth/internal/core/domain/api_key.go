package domain

import (
	"crypto/rand"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/base58"
)

const (
	byteLength = 16
)

// ApiKey represents an identification token used to authenticate and
// authorize specific endpoints with rate limit constraints.
type ApiKey struct {
	ID        string     // ID is a unique identifier for the API key.
	Service   string     // Service that this API key can access.
	RateLimit *RateLimit // RateLimit defines the request constraints over a specific time range for this API key.
	IsRoot    bool       // IsRoot specifies whether this API key is a root key. Root key can access all endpoints.
}

// RateLimit defines the number of requests allowed in a specific time range.
type RateLimit struct {
	RequestsPerRange int // RequestsPerRange is the maximum number of requests allowed within the specified time range.
	RangeInSeconds   int // RangeInSeconds specifies the duration of the time range, in seconds, for rate limiting.
}

func NewApiKey(
	service string, limit RateLimit,
) (*ApiKey, error) {
	key, err := genKey()
	if err != nil {
		return nil, err
	}

	rLimit := new(RateLimit)
	if limit.RequestsPerRange > 0 {
		if limit.RangeInSeconds <= 0 {
			return nil, fmt.Errorf("invalid rate limit range")
		}

		rLimit.RequestsPerRange = limit.RequestsPerRange
		rLimit.RangeInSeconds = limit.RangeInSeconds
	}

	apiKey := &ApiKey{
		ID:        key,
		Service:   service,
		RateLimit: rLimit,
	}

	return apiKey, nil
}

func NewRootApiKey(rootID string) *ApiKey {
	return &ApiKey{
		ID:     rootID,
		IsRoot: true,
	}
}

func genKey() (string, error) {
	random := make([]byte, byteLength)
	_, err := rand.Read(random)
	if err != nil {
		return "", fmt.Errorf("unable to read random data")
	}

	key := base58.Encode(random)
	return key, nil
}
