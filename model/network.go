package model

import (
	"fmt"
	"regexp"
	"time"
)

// Network structure
type Network struct {
	Id          string     `json:"id"          bson:"id"`
	AccountID   string     `json:"accountid"   bson:"accountid"`
	NetName     string     `json:"netName"     bson:"netName"`
	Description string     `json:"description" bson:"description"`
	Tags        []string   `json:"tags"        bson:"tags"`
	CreatedBy   string     `json:"createdBy"   bson:"createdBy"`
	UpdatedBy   string     `json:"updatedBy"   bson:"updatedBy"`
	Created     *time.Time `json:"created"     bson:"created"`
	Updated     *time.Time `json:"updated"     bson:"updated"`
	ForceUpdate bool       `json:"forceUpdate" bson:"forceUpdate"`
	Critical    bool       `json:"critical"    bson:"critical"`
	Policies    Policies   `json:"policies"    bson:"policies"`
	Default     *Settings  `json:"default"     bson:"default"`
}

type Policies struct {
	UserEndpoints bool `json:"userEndpoints" bson:"userEndpoints"`
	OnlyEndpoints bool `json:"onlyEndpoints" bson:"onlyEndpoints"`
}

// IsValid check if model is valid
func (a Network) IsValid() []error {
	errs := make([]error, 0)

	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	// check if the name empty
	if a.NetName == "" {
		errs = append(errs, fmt.Errorf("netName is required"))
	}
	// check the name field is between 3 to 40 chars
	if len(a.NetName) < 2 || len(a.NetName) > 12 {
		errs = append(errs, fmt.Errorf("name field must be between 2-12 chars"))
	}

	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.NetName)

	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("name field can only contain ascii chars a-z, 0-9"))
	}

	return errs
}
