package model

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nettica-com/nettica-admin/util"
)

// VPN structure
type VPN struct {
	Id        string    `json:"id"                        bson:"id"`
	AccountId string    `json:"accountid"                 bson:"accountid"`
	DeviceId  string    `json:"deviceid"                  bson:"deviceid"`
	Name      string    `json:"name"                      bson:"name"`
	NetId     string    `json:"netid"                     bson:"netid"`
	NetName   string    `json:"netName"                   bson:"netName"`
	Role      string    `json:"role"                      bson:"role"`
	Enable    bool      `json:"enable"                    bson:"enable"`
	Tags      []string  `json:"tags"                      bson:"tags"`
	CreatedBy string    `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy string    `json:"updatedBy"                 bson:"updatedBy"`
	Created   time.Time `json:"created"                   bson:"created"`
	Updated   time.Time `json:"updated"                   bson:"updated"`
	Current   Settings  `json:"current"                   bson:"current"`
	Default   Settings  `json:"default"                   bson:"default"`
}

type Device struct {
	Id        string    `json:"id"                        bson:"id"`
	AccountId string    `json:"accountid"                 bson:"accountid"`
	ApiKey    string    `json:"apiKey"                    bson:"apiKey"`
	Name      string    `json:"name"                      bson:"name"`
	Enable    bool      `json:"enable"                    bson:"enable"`
	Tags      []string  `json:"tags"                      bson:"tags"`
	Platform  string    `json:"platform"                  bson:"platform"`
	Version   string    `json:"version"                   bson:"version"`
	CreatedBy string    `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy string    `json:"updatedBy"                 bson:"updatedBy"`
	Created   time.Time `json:"created"                   bson:"created"`
	Updated   time.Time `json:"updated"                   bson:"updated"`
	LastSeen  time.Time `json:"lastSeen"                  bson:"lastSeen"`
	Vpns      []VPN     `json:"vpns"                      bson:"vpns"`
}

// IsValid check if model is valid
func (a VPN) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	// check the name field is between 3 to 40 chars
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
	// check if the address are valid
	for _, address := range a.Current.Address {
		if !util.IsValidCidr(address) {
			errs = append(errs, fmt.Errorf("address %s is invalid", address))
		}
	}

	return errs
}
