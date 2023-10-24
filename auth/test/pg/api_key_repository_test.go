package pgtest

import "prem-gateway/auth/internal/core/domain"

func (p *PgDbTestSuite) TestApiKeyRepository() {
	rootApiKey := domain.NewRootApiKey("rootKey")

	err := repositorySvc.ApiKeyRepository().CreateApiKey(ctx, *rootApiKey)
	p.NoError(err)

	keys, err := repositorySvc.ApiKeyRepository().GetAllApiKeys(ctx)
	p.NoError(err)

	p.Equal(1, len(keys))
	p.Equal(rootApiKey.ID, keys[0].ID)
	p.Equal(true, keys[0].IsRoot)
	p.Nil(keys[0].RateLimit)
	p.Equal("", keys[0].Service)

	apiKey, err := domain.NewApiKey(
		"mistral",
		domain.RateLimit{
			RequestsPerRange: 10,
			RangeInSeconds:   60,
		},
	)
	p.NoError(err)
	err = repositorySvc.ApiKeyRepository().CreateApiKey(ctx, *apiKey)
	keyID1 := apiKey.ID

	apiKey, err = domain.NewApiKey(
		"vicuna",
		domain.RateLimit{
			RequestsPerRange: 5,
			RangeInSeconds:   120,
		},
	)
	p.NoError(err)
	err = repositorySvc.ApiKeyRepository().CreateApiKey(ctx, *apiKey)
	keyID2 := apiKey.ID

	keys, err = repositorySvc.ApiKeyRepository().GetAllApiKeys(ctx)
	p.NoError(err)
	p.Equal(3, len(keys))

	for _, v := range keys {
		if v.ID == keyID1 {
			p.Equal(false, v.IsRoot)
			p.Equal(10, v.RateLimit.RequestsPerRange)
			p.Equal(60, v.RateLimit.RangeInSeconds)
			p.Equal("mistral", v.Service)
		} else if v.ID == keyID2 {
			p.Equal(false, v.IsRoot)
			p.Equal(5, v.RateLimit.RequestsPerRange)
			p.Equal(120, v.RateLimit.RangeInSeconds)
			p.Equal("vicuna", v.Service)
		}
	}

	//check unique constraint on service name
	apiKey, err = domain.NewApiKey(
		"mistral",
		domain.RateLimit{
			RequestsPerRange: 10,
			RangeInSeconds:   60,
		},
	)
	p.NoError(err)

	err = repositorySvc.ApiKeyRepository().CreateApiKey(ctx, *apiKey)
	p.Equal(domain.ErrApiKeyExistForService, err)

	apiKey, err = repositorySvc.ApiKeyRepository().GetServiceApiKey(ctx, "mistral")
	p.NoError(err)
	p.Equal(keyID1, apiKey.ID)
	p.Equal(false, apiKey.IsRoot)
	p.Equal(10, apiKey.RateLimit.RequestsPerRange)
	p.Equal(60, apiKey.RateLimit.RangeInSeconds)
	p.Equal("mistral", apiKey.Service)

	apiKey, err = repositorySvc.ApiKeyRepository().GetServiceApiKey(ctx, "dummy")
	p.Equal(domain.ErrApiKeyNotFound, err)

	rapk, err := repositorySvc.ApiKeyRepository().GetRootApiKey(ctx)
	p.NoError(err)
	p.Equal(rootApiKey.ID, rapk)
}
