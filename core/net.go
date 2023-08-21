package core

import (
	"errors"
	"reflect"
	"sort"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CreateNet net with all necessary data
func CreateNet(net *model.Network) (*model.Network, error) {

	var err error
	net.Id, err = util.RandomString(12)
	if err != nil {
		return nil, err
	}
	net.Id = "net-" + net.Id

	ips := make([]string, 0)
	// normalize ip addresses given
	for _, network := range net.Default.Address {
		ip, err := util.GetNetworkAddress(network)
		if err != nil {
			return nil, err
		}
		if util.IsIPv6(ip) {
			ip = ip + "/64"
		} else {
			ip = ip + "/24"
		}
		ips = append(ips, ip)
	}

	net.Default.Address = ips
	if len(net.Default.AllowedIPs) == 0 {
		net.Default.AllowedIPs = ips
	}

	net.Created = time.Now().UTC()
	net.Updated = net.Created

	if net.Default.PresharedKey == "" {
		presharedKey, err := wgtypes.GenerateKey()
		if err != nil {
			return nil, err
		}
		net.Default.PresharedKey = presharedKey.String()
	}

	// check if net is valid
	errs := net.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("net validation error")
		}
		return nil, errors.New("failed to validate net")
	}

	err = mongo.Serialize(net.Id, "id", "networks", net)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(net.Id, "id", "networks", reflect.TypeOf(model.Network{}))
	if err != nil {
		return nil, err
	}
	net = v.(*model.Network)

	// data modified, dump new config
	return net, nil
}

// ReadNet net by id
func ReadNet(id string) (*model.Network, error) {
	v, err := mongo.Deserialize(id, "id", "networks", reflect.TypeOf(model.Network{}))
	if err != nil {
		return nil, err
	}
	net := v.(*model.Network)

	return net, nil
}

// UpdateNet preserve keys
func UpdateNet(Id string, net *model.Network) (*model.Network, error) {
	v, err := mongo.Deserialize(Id, "id", "networks", reflect.TypeOf(model.Network{}))
	if err != nil {
		return nil, err
	}
	//	current := v.(*model.Network)

	if v == nil {
		return nil, errors.New("net is nil")
		//		x: = fmt.Sprintf("could not retrieve net %s", Id)
		//		return nil, errors.New(x)
	}

	//	if current.ID != Id {
	//		return nil, errors.New("records Id mismatch")
	//	}

	// check if net is valid
	errs := net.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("net validation error")
		}
		return nil, errors.New("failed to validate net")
	}

	net.Updated = time.Now().UTC()

	err = mongo.Serialize(net.Id, "id", "networks", net)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "networks", reflect.TypeOf(model.Network{}))
	if err != nil {
		return nil, err
	}
	net = v.(*model.Network)

	// data modified, dump new config
	return net, nil
}

// DeleteNet from disk
func DeleteNet(id string) error {

	err := mongo.Delete(id, "id", "networks")
	//	path := filepath.Join(os.Getenv("WG_CONF_DIR"), id)
	//	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}

// ReadNetworks all clients
func ReadNetworks(email string) ([]*model.Network, error) {

	accounts, err := mongo.ReadAllAccounts(email)

	results := make([]*model.Network, 0)

	for _, account := range accounts {
		if account.NetId != "" && account.Status == "Active" {
			nets := mongo.ReadAllNetworks("id", account.NetId)
			results = append(results, nets...)
		} else if account.Status == "Active" {
			nets := mongo.ReadAllNetworks("accountid", account.Parent)
			results = append(results, nets...)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Created.After(results[j].Created)
	})

	return results, err
}
