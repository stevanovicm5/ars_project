package model

import "github.com/google/uuid"

// Parameter represents a key-value pair within a configuration or a label.
//
//	@Description	Key-value parameter for configurations
type Parameter struct {
	//	@Description	Parameter key
	Key string `json:"key"`
	//	@Description	Parameter value
	Value string `json:"value"`
}

// Configuration represents a single configuration item.
//
//	@Description	Configuration entity with parameters
type Configuration struct {
	//	@Description	Unique identifier for the configuration
	ID uuid.UUID `json:"id"`
	//	@Description	Name of the configuration
	Name string `json:"name"`
	//	@Description	Version of the configuration
	Version string `json:"version"`
	//	@Description	List of configuration parameters
	Params []Parameter `json:"params"`
}

// ConfigurationGroup represents a collection of configurations.
//
//	@Description	Group containing multiple configurations
type ConfigurationGroup struct {
	//	@Description	Unique identifier for the group
	ID uuid.UUID `json:"id"`
	//	@Description	Name of the configuration group
	Name string `json:"name"`
	//	@Description	Version of the configuration group
	Version string `json:"version"`
	//	@Description	List of configurations in this group
	Configurations []Configuration `json:"configurations"`
}

// CreateConfigurationRequest represents the request body for creating a configuration.
//
//	@Description	Request model for creating a new configuration
type CreateConfigurationRequest struct {
	//	@Description	Name of the configuration
	Name string `json:"name"`
	//	@Description	Version of the configuration
	Version string `json:"version"`
	//	@Description	List of configuration parameters
	Params []Parameter `json:"params"`
}

// CreateGroupRequest represents the request body for creating a configuration group.
//
//	@Description	Request model for creating a new configuration group
type CreateGroupRequest struct {
	//	@Description	Name of the configuration group
	Name string `json:"name"`
	//	@Description	Version of the configuration group
	Version string `json:"version"`
	//	@Description	List of configurations to include in the group
	Configurations []Configuration `json:"configurations"` // Using full config model for simplicity here
}
