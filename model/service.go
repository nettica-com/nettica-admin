package model

import (
	"fmt"
	"time"
)

// Net structure
type Service struct {
	Id             string    `json:"id"             bson:"id"`
	ServiceGroup   string    `json:"serviceGroup"   bson:"serviceGroup"`
	ApiKey         string    `json:"apikey"         bson:"apikey"`
	AccountID      string    `json:"accountid"      bson:"accountid"`
	Email          string    `json:"email"          bson:"email"`
	SubscriptionId string    `json:"subscriptionid" bson:"subscriptionid"`
	Created        time.Time `json:"created"        bson:"created"`
	Updated        time.Time `json:"updated"        bson:"updated"`
	CreatedBy      string    `json:"createdBy"      bson:"createdBy"`
	UpdatedBy      string    `json:"updatedBy"      bson:"updatedBy"`
	Device         *Device   `json:"device"         bson:"device"`
	Net            *Network  `json:"net"            bson:"net"`
	VPN            *VPN      `json:"vpn"            bson:"vpn"`
	ContainerId    string    `json:"containerid"    bson:"containerid"`
	Status         string    `json:"status"         bson:"status"`
	ServiceHost    string    `json:"serviceHost"    bson:"serviceHost"`
	Name           string    `json:"name"           bson:"name"`
	Description    string    `json:"description"    bson:"description"`
	ServiceType    string    `json:"serviceType"    bson:"serviceType"`
	ServicePort    int       `json:"servicePort"    bson:"servicePort"`
	DefaultSubnet  string    `json:"defaultSubnet"  bson:"defaultSubnet"`
}

// IsValid check if model is valid
func (s Service) IsValid() []error {
	errs := make([]error, 0)

	if s.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	return errs
}
