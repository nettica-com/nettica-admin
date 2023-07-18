package model

import (
	"fmt"
	"time"

	"github.com/nettica-com/nettica-admin/util"
)

type Account struct {
	Id          string    `json:"id"          bson:"id"`
	Parent      string    `json:"parent"      bson:"parent"`
	MeshId      string    `json:"meshId"      bson:"meshId"`
	MeshName    string    `json:"meshName"    bson:"meshName"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	Role        string    `json:"role"        bson:"role"`
	Status      string    `json:"status"      bson:"status"`
	Key         string    `json:"key"         bson:"key"`
	From        string    `json:"from"        bson:"from"`
	AccountName string    `json:"accountName" bson:"accountName"`
	Created     time.Time `json:"created"     bson:"created"`
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
