package services

import (
	"alati_projekat/model"
	"context"
)

type Service interface {
	AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error
	GetConfiguration(ctx context.Context, name string, version string) (model.Configuration, error)
	UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (model.Configuration, error)
	DeleteConfiguration(ctx context.Context, name string, version string) error

	AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error
	GetConfigurationGroup(ctx context.Context, name string, version string) (model.ConfigurationGroup, error)
	UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (model.ConfigurationGroup, error)
	DeleteConfigurationGroup(ctx context.Context, name string, version string) error

	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	SaveIdempotencyKey(ctx context.Context, key string)

	FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) ([]model.Configuration, error)
	DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (int, error)
}
