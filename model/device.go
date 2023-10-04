package model

import (
	"fmt"
	"regexp"
	"time"
)

// Device structure
// swagger:model Device
// Defines the device model
type Device struct {
	Id            string     `json:"id"                        bson:"id"`
	Server        string     `json:"server"                    bson:"server"`
	ApiKey        string     `json:"apiKey"                    bson:"apiKey"`
	AccountID     string     `json:"accountid"                 bson:"accountid"`
	Name          string     `json:"name"                      bson:"name"`
	Description   string     `json:"description"               bson:"description"`
	Type          string     `json:"type"                      bson:"type"`
	Enable        bool       `json:"enable"                    bson:"enable"`
	Tags          []string   `json:"tags"                      bson:"tags"`
	Platform      string     `json:"platform"                  bson:"platform"`
	OS            string     `json:"os"                        bson:"os"`
	Architecture  string     `json:"arch"                      bson:"arch"`
	ClientID      string     `json:"clientid"                  bson:"clientid"`
	AuthDomain    string     `json:"authdomain"                bson:"authdomain"`
	ApiID         string     `json:"apiid"                     bson:"apiid"`
	AppData       string     `json:"appdata"                   bson:"appdata"`
	CheckInterval int64      `json:"checkInterval"             bson:"checkInterval"`
	ServiceGroup  string     `json:"serviceGroup"              bson:"serviceGroup"`
	ServiceApiKey string     `json:"serviceApiKey"             bson:"serviceApiKey"`
	SourceAddress string     `json:"sourceAddress"             bson:"sourceAddress"`
	Debug         bool       `json:"debug"                     bson:"debug"`
	Quiet         bool       `json:"quiet"                     bson:"quiet"`
	Authenticated bool       `json:"authenticated"             bson:"authenticated"`
	Version       string     `json:"version"                   bson:"version"`
	InstanceID    string     `json:"instanceid"                bson:"instanceid"`
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

	// check the name field is between 2 to 40 chars
	if len(a.Name) < 2 || len(a.Name) > 40 {
		errs = append(errs, fmt.Errorf("name field must be between 2-40 chars"))
	}
	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.Name)

	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("name field can only contain ascii chars a-z,-,0-9"))
	}

	if a.Server == "" {
		errs = append(errs, fmt.Errorf("server field is required"))
	}

	if a.ApiKey == "" {
		errs = append(errs, fmt.Errorf("apiKey field is required"))
	}

	return errs
}
