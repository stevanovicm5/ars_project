package services

import (
	"alati_projekat/model"
)

type Service interface {
	// Configurations
	AddConfiguration(config model.Configuration, idempotencyKey string) error
	GetConfiguration(name string, version string) (model.Configuration, error)
	UpdateConfiguration(config model.Configuration, idempotencyKey string) error
	DeleteConfiguration(name string, version string) error

	// Group configurations
	AddConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) error
	GetConfigurationGroup(name string, version string) (model.ConfigurationGroup, error)
	UpdateConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) error
	DeleteConfigurationGroup(name string, version string) error

	// Idempotency
	CheckIdempotencyKey(key string) (bool, error)
	SaveIdempotencyKey(key string)
}
