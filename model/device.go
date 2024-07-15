package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Device structure
// swagger:model Device
// Defines the device model
type Device struct {
	Version       string     `json:"version"                   bson:"version"`
	Id            string     `json:"id"                        bson:"id"`
	Server        string     `json:"server"                    bson:"server"`
	ApiKey        string     `json:"apiKey"                    bson:"apiKey"`
	AccountID     string     `json:"accountid"                 bson:"accountid"`
	Name          string     `json:"name"                      bson:"name"`
	Description   string     `json:"description"               bson:"description"`
	Type          string     `json:"type,omitempty"            bson:"type,omitempty"`
	Enable        bool       `json:"enable"                    bson:"enable"`
	Tags          []string   `json:"tags"                      bson:"tags"`
	Platform      string     `json:"platform"                  bson:"platform"`
	OS            string     `json:"os"                        bson:"os"`
	Architecture  string     `json:"arch"                      bson:"arch"`
	CheckInterval int64      `json:"checkInterval"             bson:"checkInterval"`
	ServiceGroup  string     `json:"serviceGroup,omitempty"    bson:"serviceGroup,omitempty"`
	ServiceApiKey string     `json:"serviceApiKey,omitempty"   bson:"serviceApiKey,omitempty"`
	SourceAddress string     `json:"sourceAddress,omitempty"   bson:"sourceAddress,omitempty"`
	Logging       string     `json:"logging"                   bson:"logging"`
	Registered    bool       `json:"registered"                bson:"registered"`
	UpdateKeys    bool       `json:"updateKeys"                bson:"updateKeys"`
	InstanceID    string     `json:"instanceid,omitempty"      bson:"instanceid,omitempty"`
	EZCode        string     `json:"ezcode,omitempty"          bson:"ezcode,omitempty"`
	CreatedBy     string     `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy     string     `json:"updatedBy"                 bson:"updatedBy"`
	Created       time.Time  `json:"created"                   bson:"created"`
	Updated       time.Time  `json:"updated"                   bson:"updated"`
	LastSeen      *time.Time `json:"lastSeen,omitempty"        bson:"lastSeen,omitempty"`
	VPNs          []*VPN     `json:"vpns,omitempty"            bson:"vpns,omitempty"`
}

// IsValid check if model is valid
func (a Device) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	if a.AccountID == "" {
		errs = append(errs, fmt.Errorf("accountid is required"))
	}

	// check the name field is between 1 to 253 chars
	if len(a.Name) < 1 || len(a.Name) > 253 {
		errs = append(errs, fmt.Errorf("name field must be between 2-40 chars"))
	}
	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.Name)
	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("name field can only contain ascii chars a-z,-,0-9"))
	}
	parts := strings.Split(a.Name, ".")
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) < 1 || len(parts[i]) > 63 {
			errs = append(errs, fmt.Errorf("each name field must be between 1-63 chars"))
		}
	}

	if a.Server == "" {
		errs = append(errs, fmt.Errorf("server field is required"))
	}

	if a.ApiKey == "" {
		errs = append(errs, fmt.Errorf("apiKey field is required"))
	}

	return errs
}
