package domain

import "context"

type ApiKeyRepository interface {
	// CreateApiKey creates a new API key.
	CreateApiKey(ctx context.Context, key ApiKey) error
	// GetApiKey returns the API key with the given ID.
	GetApiKey(ctx context.Context, id string) (*ApiKey, error)
	// DeleteApiKey deletes the API key with the given ID.
	DeleteApiKey(ctx context.Context, id string) error
	// GetAllApiKeys returns all API keys.
	GetAllApiKeys(ctx context.Context) ([]ApiKey, error)
	// GetServiceApiKey returns the API key for the given service.
	GetServiceApiKey(ctx context.Context, service string) (*ApiKey, error)
	//GetRootApiKey returns the root API key.
	GetRootApiKey(ctx context.Context) (string, error)
}
