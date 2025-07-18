package model

import (
	"fmt"
	"time"
)

type PurchaseRceipt struct {
	AccountID string `json:"accountid" bson:"accountid"`
	Email     string `json:"email"     bson:"email"`
	Name      string `json:"name"      bson:"name"`
	Source    string `json:"source"    bson:"source"`
	ProductID string `json:"productid" bson:"productid"`
	Receipt   string `json:"receipt"   bson:"receipt"`
}

// Subscription structure
type Subscription struct {
	Id          string     `json:"id"                  bson:"id"`
	AccountID   string     `json:"accountid"           bson:"accountid"`
	Email       string     `json:"email"               bson:"email"`
	Name        string     `json:"name"                bson:"name"`
	Description string     `json:"description"         bson:"description"`
	Issued      *time.Time `json:"issued"              bson:"issued"`
	Expires     *time.Time `json:"expires"             bson:"expires"`
	LastUpdated *time.Time `json:"lastUpdated"         bson:"lastUpdated"`
	CreatedBy   string     `json:"createdBy"           bson:"createdBy"`
	UpdatedBy   string     `json:"updatedBy"           bson:"updatedBy"`
	Status      string     `json:"status"              bson:"status"`
	Sku         string     `json:"sku"                 bson:"sku"`
	Credits     int        `json:"credits"             bson:"credits"`
	AutoRenew   bool       `json:"autoRenew"           bson:"autoRenew"`
	Original    string     `json:"original,omitempty"  bson:"original,omitempty"`
	TxIds       string     `json:"txIds,omitempty"     bson:"txIds,omitempty"`
	Receipt     string     `json:"receipt,omitempty"   bson:"receipt,omitempty"`
	IsDeleted   *bool      `json:"isDeleted,omitempty" bson:"isDeleted,omitempty"`
}

// IsValid check if model is valid
func (s Subscription) IsValid() []error {
	errs := make([]error, 0)

	if s.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	return errs
}
