package core

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

var (
	CreateLock sync.Mutex
)

// CreateService service with all necessary data
func CreateService(service *model.Service) (*model.Service, error) {

	// lock the function so only one service can be created at a time
	CreateLock.Lock()
	defer CreateLock.Unlock()

	var err error
	u := uuid.NewV4()
	service.Id = u.String()
	service.Created = time.Now().UTC()
	service.Updated = time.Now().UTC()

	if service.ApiKey == "" {
		service.ApiKey, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
	}

	// TODO: validate the subscription

	// find the account id for this user
	accounts, err := ReadAllAccounts(service.CreatedBy)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.Parent == account.Id {
			service.AccountId = account.Id
			break
		}
	}

	if service.ServicePort == 0 {
		service.ServicePort = 30000
	}
	if service.DefaultSubnet == "" {
		service.DefaultSubnet = "10.10.10.0/24"
	}

	if service.Relay.NetId == "" {
		// get all the current nets and see if there is one with the same name
		nets, err := ReadNetworks(service.CreatedBy)
		if err != nil {
			return nil, err
		}

		found := false

		for _, m := range nets {
			if m.NetName == service.Relay.NetName {
				found = true
				service.Relay.NetName = m.NetName
				service.Relay.NetId = m.Id
				service.Relay.Default = m.Default
				break
			}
		}

		if !found {
			// create a default net
			net := model.Network{
				AccountId:   service.AccountId,
				NetName:     service.Relay.NetName,
				Description: service.Description,
				Created:     time.Now().UTC(),
				Updated:     time.Now().UTC(),
				CreatedBy:   service.CreatedBy,
			}
			net.Default.Address = []string{service.DefaultSubnet}
			net.Default.Dns = service.Relay.Current.Dns
			net.Default.EnableDns = false
			net.Default.UPnP = false

			net2, err := CreateNet(&net)
			if err != nil {
				return nil, err
			}
			service.Relay.NetName = net2.NetName
			service.Relay.NetId = net2.Id
			service.Relay.Default = net2.Default
		}
	} else {
		// check if net exists
		net, err := ReadNet(service.Relay.NetId)
		if err != nil {
			return nil, err
		}
		if net == nil {
			return nil, errors.New("net does not exist")
		}
		service.Relay.NetName = net.NetName
		service.Relay.NetId = net.Id
		service.Relay.Default = net.Default
		log.Infof("Using existing net: %s", net.NetName)
	}

	if service.Relay.Id == "" {
		// create a default vpn using the net
		vpn := model.VPN{
			Id:        uuid.NewV4().String(),
			AccountId: service.AccountId,
			Name:      strings.ToLower(service.ServiceType) + "." + service.Relay.NetName,
			Enable:    true,
			NetId:     service.Relay.NetId,
			NetName:   service.Relay.NetName,
			//			VPNGroup: service.Relay.VPNGroup,
			Current: service.Relay.Current,
			Default: service.Relay.Default,
			//			Type:      "ServiceVPN",
			Created:   time.Now().UTC(),
			Updated:   time.Now().UTC(),
			CreatedBy: service.CreatedBy,
		}

		// Failsafe entry for DNS.  Service will break without proper DNS setup.  If nothing is set use google
		if len(vpn.Current.Dns) == 0 {
			vpn.Current.Dns = append(vpn.Current.Dns, "8.8.8.8")
		}

		// Configure the routing for the relay/egress vpn
		if vpn.Current.PostUp == "" {
			vpn.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
		}
		if vpn.Current.PostDown == "" {
			vpn.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", vpn.NetName, vpn.NetName)
		}

		vpn.Current.PersistentKeepalive = 23

		switch service.ServiceType {
		case "Relay":
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)

		case "Tunnel":
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")

		case "Ingress":
			vpn.Role = "Ingress"
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")

		case "Egress":
			vpn.Role = "Egress"
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")

		}

		vpn2, err := CreateVPN(&vpn)
		if err != nil {
			return nil, err
		}
		service.Relay = *vpn2
	}

	// check if service is valid
	errs := service.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.Error(err)
		}
		return nil, errors.New("failed to validate service")
	}

	// create the service
	err = mongo.Serialize(service.Id, "id", "service", service)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(service.Id, "id", "service", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	service = v.(*model.Service)

	// return the service
	return service, nil
}

// ReadService service by id
func ReadService(id string) (*model.Service, error) {
	v, err := mongo.Deserialize(id, "id", "service", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	service := v.(*model.Service)

	return service, nil
}

// UpdateService preserve keys
func UpdateService(Id string, service *model.Service) (*model.Service, error) {
	v, err := mongo.Deserialize(Id, "id", "service", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.Service)

	if v == nil {
		return nil, errors.New("service is nil")
		//		x: = fmt.Sprintf("could not retrieve service %s", Id)
		//		return nil, errors.New(x)
	}

	if current.Id != Id {
		return nil, errors.New("records Id mismatch")
	}

	// check if service is valid
	errs := service.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("service validation error")
		}
		return nil, errors.New("failed to validate service")
	}

	service.Updated = time.Now().UTC()

	err = mongo.Serialize(service.Id, "id", "service", service)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "service", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	service = v.(*model.Service)

	// data modified, dump new config
	return service, nil
}

// DeleteService from database
func DeleteService(id string) error {

	// Get the service
	v, err := mongo.Deserialize(id, "id", "service", reflect.TypeOf(model.Service{}))
	if err != nil {
		log.Errorf("failed to delete service %s", id)
		return err
	}
	service := v.(*model.Service)

	if service.Relay.Id != "" {
		err = DeleteVPN(service.Relay.Id)
		if err != nil {
			log.Errorf("failed to delete vpn %s", service.Relay.Id)
			return err
		}
	}

	if service.Relay.NetId != "" {
		vpns, err := ReadVPN2("netid", service.Relay.NetId)
		if err != nil {
			log.Errorf("failed to delete net %s", service.Relay.NetId)
			return err
		}
		if len(vpns) == 0 {
			err = DeleteNet(service.Relay.NetId)
			if err != nil {
				log.Errorf("failed to delete net %s", service.Relay.NetId)
				return err
			}
		}
	}

	// Now delete the service

	err = mongo.Delete(id, "id", "service")
	if err != nil {
		return err
	}

	return nil
}

// ReadServices all clients
func ReadServices(email string) ([]*model.Service, error) {

	accounts, err := mongo.ReadAllAccounts(email)

	results := make([]*model.Service, 0)

	for _, account := range accounts {
		if account.Id == account.Parent && account.Status == "Active" {
			services, err := mongo.ReadAllServices(email)
			if err == nil {
				results = append(results, services...)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Created.After(results[j].Created)
	})

	return results, err
}

// ReadServiceVPN returns all services configured for a vpn
func ReadServiceVPN(serviceGroup string) ([]*model.Service, error) {
	services, err := mongo.ReadServiceHost(serviceGroup)
	return services, err
}
