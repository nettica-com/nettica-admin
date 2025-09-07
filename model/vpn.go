package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// EnableDisableVPN structure
// This is used to enable or disable a VPN without changing any other settings
// It contains only the fields required for authentication and identification
type EnableDisableVPN struct {
	Id        string `json:"id"                        bson:"id"`
	AccountID string `json:"accountid"                 bson:"accountid"`
	DeviceID  string `json:"deviceid"                  bson:"deviceid"`
}

// VPN structure
type VPN struct {
	Id        string     `json:"id"                        bson:"id"`
	AccountID string     `json:"accountid"                 bson:"accountid"`
	DeviceID  string     `json:"deviceid"                  bson:"deviceid"`
	Name      string     `json:"name"                      bson:"name"`
	NetId     string     `json:"netid"                     bson:"netid"`
	NetName   string     `json:"netName"                   bson:"netName"`
	Role      string     `json:"role,omitempty"            bson:"role,omitempty"`
	Type      string     `json:"type,omitempty"            bson:"type,omitempty"`
	Failover  int        `json:"failover"                  bson:"failover"`
	FailCount int        `json:"failCount"                 bson:"failCount"`
	Enable    bool       `json:"enable"                    bson:"enable"`
	ReadOnly  *bool      `json:"readonly,omitempty"        bson:"readonly,omitempty"`
	Tags      []string   `json:"tags"                      bson:"tags"`
	Complete  *bool      `json:"complete,omitempty"        bson:"-"` // complete indicates this is a complete record
	CreatedBy string     `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy string     `json:"updatedBy"                 bson:"updatedBy"`
	Created   *time.Time `json:"created"                   bson:"created"`
	Updated   *time.Time `json:"updated"                   bson:"updated"`
	Current   *Settings  `json:"current,omitempty"         bson:"current,omitempty"`
	Default   *Settings  `json:"default,omitempty"         bson:"default,omitempty"`
	Devices   []*Device  `json:"devices,omitempty"         bson:"devices,omitempty"`
}

// IsValid check if model is valid
func (a VPN) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	// check the name field is between 1 to 253 chars
	if len(a.Name) < 1 || len(a.Name) > 253 {
		errs = append(errs, fmt.Errorf("name field must be between 1-253 chars"))
	}

	parts := strings.Split(a.Name, ".")
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) < 1 || len(parts[i]) > 63 {
			errs = append(errs, fmt.Errorf("each name field must be between 1-63 chars"))
		}
	}

	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.Name)
	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("name field can only contain ascii chars a-z,-,0-9"))
	}

	if len(a.NetName) < 2 || len(a.NetName) > 15 {
		errs = append(errs, fmt.Errorf("netName field must be between 2-15 chars"))
	}
	match, err = regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`, a.NetName)

	if !match {
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, fmt.Errorf("netName field can only contain ascii chars a-z,-,0-9"))
	}

	/*	// check if the allowedIPs empty
		if len(a.AllowedIPs) == 0 {
			errs = append(errs, fmt.Errorf("allowedIPs field is required"))
		}
		// check if the allowedIPs are valid
		for _, allowedIP := range a.AllowedIPs {
			if !util.IsValidCidr(allowedIP) {
				errs = append(errs, fmt.Errorf("allowedIP %s is invalid", allowedIP))
			}
		}
	*/ // check if the address empty

	if len(a.Current.Address) == 0 {
		errs = append(errs, fmt.Errorf("address field is required"))
	}

	return errs
}
