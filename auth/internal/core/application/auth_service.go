package application

import (
	"context"
	"prem-gateway/auth/internal/core/domain"
)

type AuthService interface {
	// AuthAdmin authenticates the admin user and returns a Root Api Key
	AuthAdmin(ctx context.Context, user, pass string) (string, error)
}

func NewAuthService(
	adminUser string, adminPass string, repositorySvc domain.RepositoryService,
) AuthService {
	return &authService{
		adminUser:     adminUser,
		adminPass:     adminPass,
		repositorySvc: repositorySvc,
	}
}

type authService struct {
	adminUser     string
	adminPass     string
	repositorySvc domain.RepositoryService
}

func (a authService) AuthAdmin(
	ctx context.Context, user, pass string,
) (string, error) {
	if user != a.adminUser || pass != a.adminPass {
		return "", ErrUnauthorized
	}

	return a.repositorySvc.ApiKeyRepository().GetRootApiKey(ctx)
}
