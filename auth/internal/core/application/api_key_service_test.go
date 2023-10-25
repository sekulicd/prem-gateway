package application

import (
	"context"
	"prem-gateway/auth/internal/core/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx = context.Background()
)

func TestCreateApiKey(t *testing.T) {
	apiKeyRepo := new(domain.MockApiKeyRepository)
	keys := []domain.ApiKey{
		{
			ID:     "root",
			IsRoot: true,
		},
	}
	apiKeyRepo.On("GetAllApiKeys", mock.Anything).Return(keys, nil)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	repo := new(domain.MockRepositoryService)
	repo.On("ApiKeyRepository").Return(apiKeyRepo)

	service, _ := NewApiKeyService(ctx, "root", repo)

	keyReq := CreateApiKeyReq{
		Service:          "service1",
		RequestsPerRange: 5,
		RangeInSeconds:   10,
	}

	id, err := service.CreateApiKey(context.Background(), "root", keyReq)
	assert.NotNil(t, id)
	assert.Nil(t, err)
}

func TestAllowRequest(t *testing.T) {
	apiKeyRepo := new(domain.MockApiKeyRepository)
	keys := []domain.ApiKey{
		{
			ID:     "root",
			IsRoot: true,
		},
		{
			ID:      "test-key",
			Service: "test-service",
			IsRoot:  false,
			RateLimit: &domain.RateLimit{
				RequestsPerRange: 5,
				RangeInSeconds:   10,
			},
		},
	}
	apiKeyRepo.On("GetAllApiKeys", mock.Anything).Return(keys, nil)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	repo := new(domain.MockRepositoryService)
	repo.On("ApiKeyRepository").Return(apiKeyRepo)

	service, _ := NewApiKeyService(ctx, "root", repo)

	// Valid key and path
	err := service.AllowRequest("test-key", "test-service")
	assert.Nil(t, err)

	// Invalid key
	err = service.AllowRequest("invalid-key", "test-service")
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorizedPath, err)

	// Invalid path for a valid key
	err = service.AllowRequest("test-key", "invalid-service")
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorizedPath, err)
}

func TestGetServiceApiKey(t *testing.T) {
	apiKeyRepo := new(domain.MockApiKeyRepository)
	keys := []domain.ApiKey{
		{
			ID:     "root",
			IsRoot: true,
		},
	}
	apiKeyRepo.On("GetAllApiKeys", mock.Anything).Return(keys, nil)
	testServiceKey := &domain.ApiKey{ID: "service-key"}
	apiKeyRepo.On("GetServiceApiKey", mock.Anything, "test-service").Return(testServiceKey, nil)
	repo := new(domain.MockRepositoryService)
	repo.On("ApiKeyRepository").Return(apiKeyRepo)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	service, _ := NewApiKeyService(ctx, "root", repo)

	keyID, err := service.GetServiceApiKey(context.Background(), "root", "test-service")

	assert.Equal(t, "service-key", keyID)
	assert.Nil(t, err)
}

func TestRateLimit(t *testing.T) {
	apiKeyRepo := new(domain.MockApiKeyRepository)
	keys := []domain.ApiKey{
		{
			ID:      "rate-limit-key",
			Service: "test-service",
			IsRoot:  false,
			RateLimit: &domain.RateLimit{
				RequestsPerRange: 2,
				RangeInSeconds:   5,
			},
		},
	}
	apiKeyRepo.On("GetAllApiKeys", mock.Anything).Return(keys, nil)
	repo := new(domain.MockRepositoryService)
	repo.On("ApiKeyRepository").Return(apiKeyRepo)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	service, _ := NewApiKeyService(ctx, "root", repo)

	err := service.AllowRequest("rate-limit-key", "test-service")
	assert.Nil(t, err)

	err = service.AllowRequest("rate-limit-key", "test-service")
	assert.Nil(t, err)

	// Exceeding the rate limit
	err = service.AllowRequest("rate-limit-key", "test-service")
	assert.NotNil(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)

	time.Sleep(6 * time.Second)
	// Rate limit should reset after the range
	err = service.AllowRequest("rate-limit-key", "test-service")
	assert.Nil(t, err)
}

func TestRequestCount(t *testing.T) {
	apiKeyRepo := new(domain.MockApiKeyRepository)
	keys := []domain.ApiKey{
		{
			ID:     "root",
			IsRoot: true,
		},
	}
	apiKeyRepo.On("GetAllApiKeys", mock.Anything).Return(keys, nil)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	repo := new(domain.MockRepositoryService)
	repo.On("ApiKeyRepository").Return(apiKeyRepo)
	apiKeyRepo.On("CreateApiKey", mock.Anything, mock.Anything).Return(nil)
	service, err := NewApiKeyService(ctx, "root", repo)
	if err != nil {
		t.Fatalf("Error initializing the service: %v", err)
	}

	// Create a new API key
	keyReq := CreateApiKeyReq{
		Service:          "test",
		RequestsPerRange: 5,
		RangeInSeconds:   3,
	}
	apiKey, err := service.CreateApiKey(context.Background(), "root", keyReq)
	assert.NoError(t, err, "Error creating API key")

	// Use the key to its limit
	for i := 0; i < 5; i++ {
		err = service.AllowRequest(apiKey, "test")
		assert.Nil(t, err)
	}

	// This request should be denied, as it exceeds the limit
	err = service.AllowRequest(apiKey, "test")
	assert.NotNil(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)

	// Fetch the key to check the request count
	keyInfo, exists := service.(*apiKeyService).getKey(apiKey)
	assert.True(t, exists, "API key should exist")
	assert.Equal(t, 6, keyInfo.requestCount, "Request count should not increment after limit")

	time.Sleep(3 * time.Second)

	// Now it should allow requests again
	err = service.AllowRequest(apiKey, "test")
	assert.Nil(t, err)
	keyInfo, exists = service.(*apiKeyService).getKey(apiKey)
	assert.True(t, exists, "API key should exist")
	assert.Equal(t, 1, keyInfo.requestCount, "Request count should not increment after limit")

	for i := 0; i < 4; i++ {
		err = service.AllowRequest(apiKey, "test")
		assert.Nil(t, err)
	}

	err = service.AllowRequest(apiKey, "test")
	assert.NotNil(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)
}
