package repository

import (
	"alati_projekat/model"
	"context"
)

// U fajlu gde je definisan repository.Repository
type Repository interface {
	// CONFIGURATIONS
	AddConfiguration(ctx context.Context, config model.Configuration) error
	GetConfiguration(ctx context.Context, name, version string) (model.Configuration, error)
	UpdateConfiguration(ctx context.Context, config model.Configuration) error
	DeleteConfiguration(ctx context.Context, name, version string) error

	// CONFIGURATION GROUPS
	AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) error
	GetConfigurationGroup(ctx context.Context, name, version string) (model.ConfigurationGroup, error)
	UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) error
	DeleteConfigurationGroup(ctx context.Context, name, version string) error

	// IDEMPOTENCY
	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	SaveIdempotencyKey(ctx context.Context, key string) error // Treba da vrati error, ne void
}
