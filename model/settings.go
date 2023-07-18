package model

import (
	"fmt"

	"github.com/nettica-com/nettica-admin/util"
)

// Host structure
type Settings struct {
	PrivateKey          string   `json:"privateKey"                bson:"privateKey"`
	PublicKey           string   `json:"publicKey"                 bson:"publicKey"`
	PresharedKey        string   `json:"presharedKey"              bson:"presharedKey"`
	AllowedIPs          []string `json:"allowedIPs"                bson:"allowedIPs"`
	Address             []string `json:"address"                   bson:"address"`
	Dns                 []string `json:"dns"                       bson:"dns"`
	Table               string   `json:"table"                     bson:"table"`
	PersistentKeepalive int      `json:"persistentKeepalive"       bson:"persistentKeepalive"`
	ListenPort          int      `json:"listenPort"                bson:"listenPort"`
	Endpoint            string   `json:"endpoint"                  bson:"endpoint"`
	Mtu                 int      `json:"mtu"                       bson:"mtu"`
	SubnetRouting       bool     `json:"subnetRouting"             bson:"subnetRouting"`
	UPnP                bool     `json:"upnp"                      bson:"upnp"`
	EnableDns           bool     `json:"enableDns"                 bson:"enableDns"`
	PreUp               string   `json:"preUp"                     bson:"preUp"`
	PostUp              string   `json:"postUp"                    bson:"postUp"`
	PreDown             string   `json:"preDown"                   bson:"preDown"`
	PostDown            string   `json:"postDown"                  bson:"postDown"`
}

// IsValid check if model is valid
func (a Settings) IsValid() []error {
	errs := make([]error, 0)

	// check if the address empty
	if len(a.Address) == 0 {
		errs = append(errs, fmt.Errorf("address field is required"))
	}
	// check if the address are valid
	for _, address := range a.Address {
		if !util.IsValidCidr(address) {
			errs = append(errs, fmt.Errorf("address %s is invalid", address))
		}
	}

	return errs
}
