package model

import (
	"fmt"
	"time"

	"github.com/nettica-com/nettica-admin/util"
)

type Account struct {
	Id             string     `json:"id"                        bson:"id"`
	Parent         string     `json:"parent"                    bson:"parent"`
	Email          string     `json:"email"                     bson:"email"`
	Sub            string     `json:"sub,omitempty"             bson:"sub,omitempty"`
	Name           string     `json:"name"                      bson:"name"`
	AccountName    string     `json:"accountName"               bson:"accountName"`
	NetId          string     `json:"netId"                     bson:"netId"`
	NetName        string     `json:"netName"                   bson:"netName"`
	Picture        string     `json:"picture"                   bson:"picture"`
	UserPicture    string     `json:"userPicture,omitempty"     bson:"userPicture,omitempty"`
	AccountPicture string     `json:"accountPicture,omitempty"  bson:"accountPicture,omitempty"`
	AccountPict    string     `json:"accountPict,omitempty"     bson:"accountPict,omitempty"`
	Role           string     `json:"role"                      bson:"role"`
	Status         string     `json:"status"                    bson:"status"`
	ApiKey         string     `json:"apiKey"                    bson:"apiKey"`
	CreatedBy      string     `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy      string     `json:"updatedBy"                 bson:"updatedBy"`
	Created        time.Time  `json:"created"                   bson:"created"`
	Updated        time.Time  `json:"updated"                   bson:"updated"`
	Networks       []*Network `json:"networks"                  bson:"networks"`
	VPNs           []*VPN     `json:"vpns"                      bson:"vpns"`
	Devices        []*Device  `json:"devices"                   bson:"devices"`
}

// IsValid check if model is valid
func (a Account) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}
	// email is required, but if provided must match regex
	if a.Email != "" {
		if !util.RegexpEmail.MatchString(a.Email) {
			errs = append(errs, fmt.Errorf("email %s is invalid", a.Email))
		}
	} else {
		errs = append(errs, fmt.Errorf("email is required"))
	}

	return errs
}
