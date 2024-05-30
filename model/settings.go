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
	Table               string   `json:"table,omitempty"           bson:"table,omitempty"`
	PersistentKeepalive int      `json:"persistentKeepalive"       bson:"persistentKeepalive"`
	ListenPort          int      `json:"listenPort"                bson:"listenPort"`
	Endpoint            string   `json:"endpoint"                  bson:"endpoint"`
	Mtu                 int      `json:"mtu"                       bson:"mtu"`
	SubnetRouting       bool     `json:"subnetRouting,omitempty"   bson:"subnetRouting,omitempty"`
	UPnP                bool     `json:"upnp,omitempty"            bson:"upnp,omitempty"`
	SyncEndpoint        bool     `json:"syncEndpoint,omitempty"    bson:"syncEndpoint,omitempty"`
	FailSafe            bool     `json:"failsafe,omitempty"        bson:"failsafe,omitempty"`
	EnableDns           bool     `json:"enableDns,omitempty"       bson:"enableDns,omitempty"`
	HasSSH              bool     `json:"hasSSH,omitempty"          bson:"hasSSH,omitempty"`
	HasRDP              bool     `json:"hasRDP,omitempty"          bson:"hasRDP,omitempty"`
	PreUp               string   `json:"preUp,omitempty"           bson:"preUp,omitempty"`
	PostUp              string   `json:"postUp,omitempty"          bson:"postUp,omitempty"`
	PreDown             string   `json:"preDown,omitempty"         bson:"preDown,omitempty"`
	PostDown            string   `json:"postDown,omitempty"        bson:"postDown,omitempty"`
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
