package services

import (
	"alati_projekat/model"
	"context"
)

type Service interface {
	// Configurations
	AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error
	GetConfiguration(ctx context.Context, name string, version string) (model.Configuration, error)
	UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error
	DeleteConfiguration(ctx context.Context, name string, version string) error

	// Group configurations
	AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error
	GetConfigurationGroup(ctx context.Context, name string, version string) (model.ConfigurationGroup, error)
	UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error
	DeleteConfigurationGroup(ctx context.Context, name string, version string) error

	// Idempotency
	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	SaveIdempotencyKey(ctx context.Context, key string)

	// Labels
	FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) ([]model.Configuration, error)
	DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (int, error)
}
