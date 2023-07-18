package model

import (
	"fmt"
	"regexp"
	"time"
)

// Mesh structure
type Mesh struct {
	Id          string    `json:"id"          bson:"id"`
	AccountId   string    `json:"accountid"   bson:"accountid"`
	MeshName    string    `json:"meshName"    bson:"meshName"`
	Description string    `json:"description" bson:"description"`
	CreatedBy   string    `json:"createdBy"   bson:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"   bson:"updatedBy"`
	Created     time.Time `json:"created"     bson:"created"`
	Updated     time.Time `json:"updated"     bson:"updated"`
	Enabled     bool      `json:"enabled"     bson:"enabled"`
	Default     Settings  `json:"default"     bson:"default"`
}

// IsValid check if model is valid
func (a Mesh) IsValid() []error {
	errs := make([]error, 0)

	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	// check if the name empty
	if a.MeshName == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	// check the name field is between 3 to 40 chars
	if len(a.MeshName) < 2 || len(a.MeshName) > 12 {
		errs = append(errs, fmt.Errorf("name field must be between 2-12 chars"))
	}

	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.MeshName)

	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("name field can only contain ascii chars a-z, 0-9"))
	}

	return errs
}
