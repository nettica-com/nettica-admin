package model

import (
	"fmt"
	"time"
)

// Limits structure defines the limits for the account

type Limits struct {
	Id          string    `json:"id"                        bson:"id"`
	AccountID   string    `json:"accountid"                 bson:"accountid"`
	Devices     int       `json:"devices"                   bson:"devices"`
	MaxDevices  int       `json:"maxdevices"                bson:"maxdevices"`
	Networks    int       `json:"networks"                  bson:"networks"`
	MaxNetworks int       `json:"maxnetworks"               bson:"maxnetworks"`
	Members     int       `json:"members"                   bson:"members"`
	MaxMembers  int       `json:"maxmembers"                bson:"maxmembers"`
	Services    int       `json:"services"                  bson:"services"`
	MaxServices int       `json:"maxservices"               bson:"maxservices"`
	Tolerance   float64   `json:"tolerance"                 bson:"tolerance"`
	CreatedBy   string    `json:"createdBy"                 bson:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"                 bson:"updatedBy"`
	Created     time.Time `json:"created"                   bson:"created"`
	Updated     time.Time `json:"updated"                   bson:"updated"`
}

// IsValid check if model is valid
func (a Limits) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	if a.AccountID == "" {
		errs = append(errs, fmt.Errorf("accountid is required"))
	}

	if a.Tolerance == 0.0 {
		errs = append(errs, fmt.Errorf("tolerance is invalid"))
	}

	return errs
}

func (l Limits) DevicesLimitReached(count int) bool {
	if l.MaxDevices < 0 {
		return false
	}

	if count >= int(float64(l.MaxDevices)*l.Tolerance) {
		return true
	}

	return false
}

func (l Limits) NetworksLimitReached(count int) bool {
	if l.MaxNetworks < 0 {
		return false
	}

	if count >= int(float64(l.MaxNetworks)*l.Tolerance) {
		return true
	}

	return false
}

func (l Limits) MembersLimitReached(count int) bool {
	if l.MaxMembers < 0 {
		return false
	}

	if count >= int(float64(l.MaxMembers)*l.Tolerance) {
		return true
	}

	return false
}

func (l Limits) RelaysLimitReached(count int) bool {
	if l.MaxServices < 0 {
		return false
	}

	if count >= int(float64(l.MaxServices)*l.Tolerance) {
		return true
	}

	return false
}
