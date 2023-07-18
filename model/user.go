package model

import (
	"fmt"
	"time"

	"github.com/nettica-com/nettica-admin/util"
)

// User structure
type User struct {
	Sub       string    `json:"sub"          bson:"sub"`
	Name      string    `json:"name"         bson:"name"`
	Email     string    `json:"email"        bson:"email"`
	AccountId string    `json:"accountid"    bson:"accountid"`
	Profile   string    `json:"profile"      bson:"profile"`
	Picture   string    `json:"picture"      bson:"picture"`
	Issuer    string    `json:"issuer"       bson:"issuer"`
	Plan      string    `json:"plan"         bson:"plan"`
	IssuedAt  time.Time `json:"issuedAt"     bson:"issuedAt"`
	CreatedBy string    `json:"createdBy"    bson:"createdBy"`
	Created   time.Time `json:"created_at"   bson:"created_at"`
	UpdatedBy string    `json:"updatedBy"    bson:"updatedBy"`
	Updated   time.Time `json:"updated_at"   bson:"updated_at"`
}

// IsValid check if model is valid
func (a User) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}

	if a.Email == "" {
		errs = append(errs, fmt.Errorf("email is required"))
	}
	// email is not required, but if provided must match regex
	if a.Email != "" {
		if !util.RegexpEmail.MatchString(a.Email) {
			errs = append(errs, fmt.Errorf("email %s is invalid", a.Email))
		}
	}

	return errs
}
