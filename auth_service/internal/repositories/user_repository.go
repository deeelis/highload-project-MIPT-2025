package repositories

import (
	"auth_service/internal/domain/models"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (string, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
