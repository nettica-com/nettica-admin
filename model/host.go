package model

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nettica-com/nettica-admin/util"
)

// Host structure
type Host struct {
	Id        string    `json:"id"                        bson:"id"`
	AccountId string    `json:"accountid"                 bson:"accountid"`
	Name      string    `json:"name"                      bson:"name"`
	HostGroup string    `json:"hostGroup"                 bson:"hostGroup"`
	Type      string    `json:"type"                      bson:"type"`
	Role      string    `json:"role"                      bson:"role"`
	NetId     string    `json:"netid"                     bson:"netid"`
	NetName   string    `json:"netName"                   bson:"netName"`
	APIKey    string    `json:"apiKey"                    bson:"apiKey"`
	Email     string    `json:"email"                     bson:"email"`
	Enable    bool      `json:"enable"                    bson:"enable"`
	Tags      []string  `json:"tags"                      bson:"tags"`
	Platform  string    `json:"platform"                  bson:"platform"`
	Version   string    `json:"version"                   bson:"version"`
	CreatedBy string    `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy string    `json:"updatedBy"                 bson:"updatedBy"`
	Created   time.Time `json:"created"                   bson:"created"`
	Updated   time.Time `json:"updated"                   bson:"updated"`
	LastSeen  time.Time `json:"lastSeen"				  bson:"lastSeen"`
	Current   Settings  `json:"current"                   bson:"current"`
	Default   Settings  `json:"default"                   bson:"default"`
}

// IsValid check if model is valid
func (a Host) IsValid() []error {
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

	// email is not required, but if provided must match regex
	if a.Email != "" {
		if !util.RegexpEmail.MatchString(a.Email) {
			errs = append(errs, fmt.Errorf("email %s is invalid", a.Email))
		}
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
