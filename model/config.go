package model

import "github.com/google/uuid"

// Parameter represents a key-value pair within a configuration or a label.
//
// @Description Key-value parameter for configurations or labels.
type Parameter struct {
	// @Description Parameter key
	Key string `json:"key"`
	// @Description Parameter value
	Value string `json:"value"`
}

// Configuration represents a single configuration item.
//
// @Description Configuration entity with parameters and labels.
type Configuration struct {
	// @Description Unique identifier for the configuration
	// @example 11a22b33-44c5-55d6-66e7-77f88g99h00i
	ID uuid.UUID `json:"id"`
	// @Description Name of the configuration
	// @example service-api
	Name string `json:"name"`
	// @Description Version of the configuration
	// @example v1
	Version string `json:"version"`
	// @Description List of configuration parameters
	Params []Parameter `json:"params"`

	// @Description List of config labels (optional)
	// @example [{"key": "env", "value": "dev"}, {"key": "region", "value": "eu"}]
	Labels []Parameter `json:"labels,omitempty"`
}

// ConfigurationGroup represents a collection of configurations.
//
// @Description Group containing multiple configurations.
type ConfigurationGroup struct {
	// @Description Unique identifier for the group
	// @example f9876e54-32d1-00c9-87b6-54a321e00f9
	ID uuid.UUID `json:"id"`
	// @Description Name of the configuration group
	// @example production-cluster
	Name string `json:"name"`
	// @Description Version of the configuration group
	// @example v2
	Version string `json:"version"`
	// @Description List of configurations in this group
	Configurations []Configuration `json:"configurations"`
}

// CreateConfigurationRequest represents the request body for creating a configuration.
//
// @Description Request model for creating a new configuration.
type CreateConfigurationRequest struct {
	// @Description Name of the configuration
	// @example service-api
	Name string `json:"name" example:"service-api"`
	// @Description Version of the configuration
	// @example v1
	Version string `json:"version" example:"v1"`
	// @Description List of configuration parameters
	Params []Parameter `json:"params"`

	// @Description Labels (k:v pairs)
	Labels []Parameter `json:"labels"`
}

// CreateGroupRequest represents the request body for creating a configuration group.
//
// @Description Request model for creating a new configuration group.
type CreateGroupRequest struct {
	// @Description Name of the configuration group
	// @example production-cluster
	Name string `json:"name" example:"production-cluster"`
	// @Description Version of the configuration group
	// @example v2
	Version string `json:"version" example:"v2"`
	// @Description List of configurations to include in the group
	Configurations []Configuration `json:"configurations"`
}

func (c Configuration) LabelsMap() map[string]string {
	m := make(map[string]string, len(c.Labels))
	for _, p := range c.Labels {
		m[p.Key] = p.Value
	}
	return m
}
