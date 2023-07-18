package model

import (
	"fmt"
)

// Server structure
type Server struct {
	Id            string `json:"id" bson:"id"`
	Name          string `json:"name" bson:"name"`
	Description   string `json:"description" bson:"description"`
	IpAddress     string `json:"ipAddress" bson:"ipAddress"`
	PortMin       int    `json:"portMin" bson:"portMin"`
	PortMax       int    `json:"portMax" bson:"portMax"`
	ServiceGroup  string `json:"serviceGroup" bson:"serviceGroup"`
	ServiceApiKey string `json:"serviceApiKey" bson:"serviceApiKey"`
	DefaultSubnet string `json:"defaultSubnet" bson:"defaultSubnet"`
}

// IsValid check if model is valid
func (a Server) IsValid() []error {
	errs := make([]error, 0)

	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	if a.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}

	if a.IpAddress == "" {
		errs = append(errs, fmt.Errorf("ipAddress is required"))
	}

	return errs
}
