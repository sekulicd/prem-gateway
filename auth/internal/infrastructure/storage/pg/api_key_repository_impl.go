package pgdb

import (
	"context"
	"database/sql"
	"github.com/jackc/pgconn"
	"prem-gateway/auth/internal/core/domain"
	"prem-gateway/auth/internal/infrastructure/storage/pg/sqlc/queries"
)

type apiKeyRepositoryImpl struct {
	querier *queries.Queries
	execTx  func(
		ctx context.Context,
		txBody func(*queries.Queries) error,
	) error
}

func NewIdentityRepositoryImpl(
	querier *queries.Queries,
	execTx func(ctx context.Context, txBody func(*queries.Queries) error) error,
) domain.ApiKeyRepository {
	return &apiKeyRepositoryImpl{
		querier: querier,
		execTx:  execTx,
	}
}

func (a *apiKeyRepositoryImpl) CreateApiKey(
	ctx context.Context, key domain.ApiKey,
) error {
	txBody := func(querierWithTx *queries.Queries) error {
		rateLimitID := sql.NullInt32{}
		if key.RateLimit != nil {
			id, err := querierWithTx.InsertRateLimitAndReturnID(
				ctx,
				queries.InsertRateLimitAndReturnIDParams{
					RequestsPerRange: sql.NullInt32{
						Int32: int32(key.RateLimit.RequestsPerRange),
						Valid: true,
					},
					RangeInSeconds: sql.NullInt32{
						Int32: int32(key.RateLimit.RangeInSeconds),
						Valid: true,
					},
				},
			)
			if err != nil {
				return err
			}

			rateLimitID.Int32 = id
			rateLimitID.Valid = true
		}

		if err := querierWithTx.InsertApiKey(
			ctx,
			queries.InsertApiKeyParams{
				ID: key.ID,
				IsRoot: sql.NullBool{
					Bool:  key.IsRoot,
					Valid: true,
				},
				RateLimitID: rateLimitID,
				ServiceName: sql.NullString{
					String: key.Service,
					Valid:  true,
				},
			},
		); err != nil {
			if pqErr, ok := err.(*pgconn.PgError); pqErr != nil &&
				ok && pqErr.Code == uniqueViolation {
				return domain.ErrApiKeyExistForService
			}
		}

		return nil
	}

	return a.execTx(ctx, txBody)
}

func (a *apiKeyRepositoryImpl) GetApiKey(
	ctx context.Context, id string,
) (*domain.ApiKey, error) {
	//TODO implement me
	panic("implement me")
}

func (a *apiKeyRepositoryImpl) DeleteApiKey(
	ctx context.Context, id string,
) error {
	//TODO implement me
	panic("implement me")
}

func (a *apiKeyRepositoryImpl) GetAllApiKeys(ctx context.Context) ([]domain.ApiKey, error) {
	apiKeysRows, err := a.querier.GetAllApiKeys(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*domain.ApiKey
	apiKeyMap := make(map[string]*domain.ApiKey)

	for _, row := range apiKeysRows {
		apk, exists := apiKeyMap[row.ID]
		if !exists {
			apk = &domain.ApiKey{
				ID:     row.ID,
				IsRoot: row.IsRoot.Bool,
			}
			if !row.IsRoot.Bool {
				apk.RateLimit = &domain.RateLimit{
					RequestsPerRange: int(row.RequestsPerRange.Int32),
					RangeInSeconds:   int(row.RangeInSeconds.Int32),
				}

				apk.Service = row.ServiceName.String
			}
			apiKeyMap[row.ID] = apk
			resp = append(resp, apk)
		}
	}

	result := make([]domain.ApiKey, len(resp))
	for i, apkPtr := range resp {
		result[i] = *apkPtr
	}

	return result, nil
}

func (a *apiKeyRepositoryImpl) GetServiceApiKey(
	ctx context.Context, serviceName string,
) (*domain.ApiKey, error) {
	apiKeysRows, err := a.querier.GetApiKeyForServiceName(ctx, sql.NullString{
		String: serviceName,
		Valid:  true,
	})
	if err != nil {
		return nil, err
	}

	if len(apiKeysRows) == 0 {
		return nil, domain.ErrApiKeyNotFound
	}

	return &domain.ApiKey{
		ID:      apiKeysRows[0].ID,
		IsRoot:  apiKeysRows[0].IsRoot.Bool,
		Service: apiKeysRows[0].ServiceName.String,
		RateLimit: &domain.RateLimit{
			RequestsPerRange: int(apiKeysRows[0].RequestsPerRange.Int32),
			RangeInSeconds:   int(apiKeysRows[0].RangeInSeconds.Int32),
		},
	}, nil
}

func (a *apiKeyRepositoryImpl) GetRootApiKey(
	ctx context.Context,
) (string, error) {
	id, err := a.querier.GetRootApiKey(ctx)
	if err != nil {
		return "", err
	}

	return id, nil
}
