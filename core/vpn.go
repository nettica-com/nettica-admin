package core

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	template "github.com/nettica-com/nettica-admin/template"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CreateVPN vpn with all necessary data
func CreateVPN(vpn *model.VPN) (*model.VPN, error) {

	var err error
	vpn.Id, err = util.RandomString(12)
	if err != nil {
		return nil, err
	}
	vpn.Id = "vpn-" + vpn.Id

	// read the nets and configure the default values
	nets, err := ReadNetworks(vpn.CreatedBy)
	if err != nil {
		return nil, err
	}

	for _, net := range nets {
		if net.NetName == vpn.NetName {
			vpn.Default = net.Default
			current := vpn.Current
			vpn.Current = net.Default
			vpn.Current.ListenPort = current.ListenPort
			vpn.Current.Endpoint = current.Endpoint
			vpn.Current.PrivateKey = current.PrivateKey
			vpn.Current.PublicKey = current.PublicKey
			vpn.Current.PostUp = current.PostUp
			vpn.Current.PostDown = current.PostDown
			vpn.Current.PersistentKeepalive = current.PersistentKeepalive
			vpn.NetId = net.Id
			vpn.AccountID = net.AccountID
			vpn.Current.AllowedIPs = current.AllowedIPs
			vpn.Current.Dns = net.Default.Dns
			if current.EnableDns {
				vpn.Current.EnableDns = current.EnableDns
			}
			if current.UPnP {
				vpn.Current.UPnP = current.UPnP
			}
			if current.SyncEndpoint && current.Endpoint != "" {
				vpn.Current.SyncEndpoint = current.SyncEndpoint
			}
			if current.HasRDP {
				vpn.Current.HasRDP = current.HasRDP
			}
			if current.HasSSH {
				vpn.Current.HasSSH = current.HasSSH
			}
		}
	}

	// if the vpn data already has a public key and empty private key,
	// we know the client has already generated a key pair
	if vpn.Current.PublicKey != "" && vpn.Current.PrivateKey == "" {
		log.Info("client has already generated a key pair")
	} else {
		// generate a new key pair
		log.Info("generating a new key pair")
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		vpn.Current.PrivateKey = key.String()
		vpn.Current.PublicKey = key.PublicKey().String()
	}

	reserverIps, err := GetAllReservedNetIps(vpn.NetId)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	ipsDns := make([]string, 0)
	for _, network := range vpn.Default.Address {
		ip, err := util.GetAvailableIp(network, reserverIps)
		if err != nil {
			return nil, err
		}
		ipsDns = append(ipsDns, ip)
		if util.IsIPv6(ip) {
			ip = ip + "/128"
		} else {
			ip = ip + "/32"
		}
		ips = append(ips, ip)
	}
	vpn.Current.Address = ips
	vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, ips...)
	if vpn.Current.EnableDns {
		device, err := ReadDevice(vpn.DeviceID)
		if err != nil {
			return nil, err
		}
		if device.OS == "darwin" {
			vpn.Current.Dns = append(vpn.Current.Dns, "127.0.0.1")
		} else {
			vpn.Current.Dns = append(vpn.Current.Dns, ipsDns[0])
		}
	}

	if vpn.Current.SubnetRouting && len(vpn.Current.PostUp) == 0 {
		vpn.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
	}
	if vpn.Current.SubnetRouting && len(vpn.Current.PostDown) == 0 {
		vpn.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
	}

	vpn.Created = time.Now().UTC()
	vpn.Updated = vpn.Created

	// check if vpn is valid
	errs := vpn.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("vpn validation error")
		}
		return nil, errors.New("failed to validate vpn")
	}

	err = mongo.Serialize(vpn.Id, "id", "vpns", vpn)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(vpn.Id, "id", "vpns", reflect.TypeOf(model.VPN{}))
	if err != nil {
		return nil, err
	}
	vpn = v.(*model.VPN)

	// data modified, dump new config
	return vpn, nil
}

// GetAllReservedIps the list of all reserved IPs, client and server
func GetAllReservedNetIps(netId string) ([]string, error) {
	clients, err := mongo.ReadAllVPNs("netid", netId)

	if err != nil {
		return nil, err
	}
	reserverIps := make([]string, 0)

	for _, client := range clients {
		if client.NetId == netId {
			for _, cidr := range client.Current.Address {
				ip, err := util.GetIpFromCidr(cidr)
				if err != nil {
					log.WithFields(log.Fields{
						"err":  err,
						"cidr": cidr,
					}).Error("failed to ip from cidr")
				} else {
					reserverIps = append(reserverIps, ip)
				}
			}
		}
	}

	return reserverIps, nil
}

// ReadVPN vpn by id
func ReadVPN(id string) (*model.VPN, error) {
	v, err := mongo.Deserialize(id, "id", "vpns", reflect.TypeOf(model.VPN{}))
	if err != nil {
		return nil, err
	}
	vpn := v.(*model.VPN)

	return vpn, nil
}

// UpdateVPN preserve keys
func UpdateVPN(Id string, vpn *model.VPN, flag bool) (*model.VPN, error) {
	v, err := mongo.Deserialize(Id, "id", "vpns", reflect.TypeOf(model.VPN{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.VPN)

	if current.Id != vpn.Id {
		return nil, errors.New("records Id mismatch")
	}

	if len(vpn.Current.Address) == 0 ||
		(len(vpn.Default.Address) > 0 && len(current.Default.Address) > 0 &&
			(vpn.Default.Address[0] != current.Default.Address[0])) {
		reserverIps, err := GetAllReservedNetIps(vpn.NetId)
		if err != nil {
			return nil, err
		}

		ips := make([]string, 0)

		for _, network := range vpn.Default.Address {
			ip, err := util.GetAvailableIp(network, reserverIps)
			if err != nil {
				return nil, err
			}
			if util.IsIPv6(ip) {
				ip = ip + "/128"
			} else {
				ip = ip + "/32"
			}
			ips = append(ips, ip)
		}
		vpn.Current.Address = ips
		vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, ips...)
	}

	if vpn.Current.EnableDns {

		// on mac the dns is at 127.0.0.1
		address := "127.0.0.1"

		device, err := ReadDevice(vpn.DeviceID)
		if err != nil {
			return nil, err
		}
		// if its not a mac its running on the vpn's ip address
		if device.OS != "darwin" {
			// Append the first address to the dns list
			address, err = util.GetIpFromCidr(vpn.Current.Address[0])
			if err != nil {
				return nil, err
			}
		}

		found := false
		for _, dns := range vpn.Current.Dns {
			if dns == address {
				found = true
				break
			}
		}
		if !found {
			vpn.Current.Dns = append(vpn.Current.Dns, address)
		}
	}

	if vpn.Current.SubnetRouting && len(vpn.Current.PostUp) == 0 {
		vpn.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
	}
	if vpn.Current.SubnetRouting && len(vpn.Current.PostDown) == 0 {
		vpn.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
	}

	// check if vpn is valid
	errs := vpn.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("vpn validation error")
		}
		return nil, errors.New("failed to validate vpn")
	}

	if !flag {
		vpn.Updated = time.Now().UTC()
	}

	err = mongo.Serialize(vpn.Id, "id", "vpns", vpn)
	if err != nil {
		return nil, err
	}

	/*
		v, err = mongo.Deserialize(Id, "id", "vpns", reflect.TypeOf(model.VPN{}))
		if err != nil {
			return nil, err
		}
		vpn = v.(*model.VPN)
	*/

	// data modified, dump new config
	return vpn, nil
}

// DeleteVPN from database
func DeleteVPN(id string) error {

	return mongo.DeleteVPN(id, "vpns")
}

// ReadVPN2 vpn by param and id
func ReadVPN2(param string, id string) ([]*model.VPN, error) {
	return mongo.ReadAllVPNs(param, id)
}

// ReadVPNs all vpns
func ReadVPNs() ([]*model.VPN, error) {
	return mongo.ReadAllVPNs("", "")
}

// ReadVPNs all vpns
func ReadVPNsForUser(email string) ([]*model.VPN, error) {
	accounts, err := mongo.ReadAllAccounts(email)
	if err != nil {
		return nil, err
	}

	nets, err := ReadNetworks(email)
	if err != nil {
		return nil, err
	}

	// put the nets into a map for easy lookup
	netMap := make(map[string]*model.Network)
	for _, net := range nets {
		netMap[net.Id] = net
	}

	results := make([]*model.VPN, 0)

	for _, account := range accounts {
		if account.Status == "Active" {
			var vpns []*model.VPN
			if account.NetId != "" {
				vpns, err = mongo.ReadAllVPNs("netid", account.NetId)
				if err != nil {
					return nil, err
				}

			} else {
				vpns, err = mongo.ReadAllVPNs("accountid", account.Parent)
				if err != nil {
					return nil, err
				}
			}

			// if the vpn is not already in the results, add it
			for _, vpn := range vpns {
				// check the network policy to see if we should show this vpn
				if account.Role == "User" || account.Role == "Guest" {
					net, ok := netMap[vpn.NetId]
					if ok {
						if net.Policies.OnlyEndpoints && vpn.Current.Endpoint == "" && vpn.CreatedBy != email {
							continue
						}
					}
				}
				found := false
				for _, result := range results {
					if result.Id == vpn.Id {
						found = true
						break
					}
				}
				if !found {
					results = append(results, vpn)
				}
			}

		}
	}

	return results, err
}

// ReadVPNConfig in wg format
func ReadVPNConfig(id string) ([]byte, *string, error) {

	netName := ""
	vpn, err := ReadVPN(id)
	if err != nil {
		return nil, nil, err
	}
	vpns, err := ReadVPN2("netid", vpn.NetId)
	if err != nil {
		return nil, nil, err
	}

	index := 0
	for j := 0; j < len(vpns); j++ {
		if vpns[j].Id == id {
			index = j
			break
		}
	}

	if index == -1 {
		log.Errorf("Error reading Net: %v", vpns)
	} else {
		vpn := vpns[index]
		netName = vpn.NetName
		vpns = append(vpns[:index], vpns[index+1:]...)

		for i := 0; i < len(vpns); i++ {
			// if the current vpn doesn't have an endpoint specified it is a client, so it does not
			// need the public keys of other clients since they can't connect to each other.  If there
			// is an endpoint specified, keep all the clients in the config.
			if vpn.Current.Endpoint == "" && vpns[i].Current.Endpoint == "" {
				vpns = append(vpns[:i], vpns[i+1:]...)
				i--
			}
		}

		config, err := template.DumpWireguardConfig(vpn, vpns)
		if err != nil {
			return nil, nil, err
		}

		return config, &netName, nil
	}
	return nil, nil, err
}
