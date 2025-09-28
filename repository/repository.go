package repository

import "alati_projekat/model"

type Repository interface {
	AddConfiguration(config model.Configuration) error
	GetConfiguration(name, version string) (model.Configuration, error)
	UpdateConfiguration(config model.Configuration) error
	DeleteConfiguration(name, version string) error

	AddConfigurationGroup(group model.ConfigurationGroup) error
	GetConfigurationGroup(name, version string) (model.ConfigurationGroup, error)
	UpdateConfigurationGroup(group model.ConfigurationGroup) error
	DeleteConfigurationGroup(name, version string) error

	CheckIdempotencyKey(key string) (bool, error)
	SaveIdempotencyKey(key string) error
}
