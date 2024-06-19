package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
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
	service.Id, err = util.RandomString(12)
	if err != nil {
		return nil, err
	}
	service.Id = "service-" + service.Id
	service.Created = time.Now().UTC()
	service.Updated = time.Now().UTC()

	if service.ApiKey == "" {
		service.ApiKey, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
		service.ApiKey = "service-api-" + service.ApiKey
	}

	// TODO: validate the subscription

	// find the account id for this user
	accounts, err := ReadAllAccounts(service.CreatedBy)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.Parent == account.Id {
			service.AccountID = account.Id
			break
		}
	}

	if service.ServicePort == 0 {
		service.ServicePort = 30001
	}
	if service.DefaultSubnet == "" {
		service.DefaultSubnet = "10.10.10.0/24"
	}

	// Find or create the network to use for the service

	if service.Net == nil {
		return nil, errors.New("net is nil")
	}

	if service.Net.NetName == "" {
		service.Net.NetName = service.Name
	}

	if service.Net.Id == "" {
		// get all the current nets and see if there is one with the same name
		nets, err := ReadNetworks(service.CreatedBy)
		if err != nil {
			return nil, err
		}

		found := false

		for _, n := range nets {
			if n.NetName == service.Net.NetName {
				found = true
				service.Net = n
				break
			}
		}

		if !found {
			var c time.Time = time.Now().UTC()
			var u time.Time = time.Now().UTC()
			// create a default net
			net := model.Network{
				AccountID:   service.AccountID,
				NetName:     service.Net.NetName,
				Description: service.Description,
				Created:     &c,
				Updated:     &u,
				CreatedBy:   service.CreatedBy,
				UpdatedBy:   service.CreatedBy,
			}
			net.Default.Address = []string{service.DefaultSubnet}
			net.Default.Dns = service.Net.Default.Dns
			net.Default.EnableDns = false
			net.Default.UPnP = false

			service.Net, err = CreateNet(&net)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// check if net exists
		net, err := ReadNet(service.Net.Id)
		if err != nil {
			log.Infof("Failed to read net: %s - it may be remote", service.Net.Id)
		} else {
			service.Net = net
		}
		if service.Net == nil {
			return nil, errors.New("net does not exist")
		}
		log.Infof("Using existing net: %s", service.Net.NetName)
	}

	// Create a device for the service container
	if service.Device == nil {

		service.Device = &model.Device{
			AccountID:   service.AccountID,
			Name:        strings.ToLower(service.ServiceType) + "." + service.Net.NetName,
			Description: service.Description,
			Enable:      true,
			Server:      os.Getenv("SERVER"),
			Type:        "Service",
			Platform:    "Linux",
			Created:     time.Now().UTC(),
			Updated:     time.Now().UTC(),
			CreatedBy:   service.CreatedBy,
			UpdatedBy:   service.CreatedBy,
		}

		// Create the device
		service.Device, err = CreateDevice(service.Device)
		if err != nil {
			return nil, err
		}
	} // otherwise it was created remotely

	if service.VPN == nil {

		// create a default vpn using the net
		var c time.Time = time.Now().UTC()
		var u time.Time = time.Now().UTC()

		vpn := model.VPN{
			AccountID: service.AccountID,
			Name:      strings.ToLower(service.ServiceType) + "." + service.Net.NetName,
			Enable:    true,
			NetId:     service.Net.Id,
			NetName:   service.Net.NetName,
			DeviceID:  service.Device.Id,
			Current:   service.VPN.Current,
			Default:   service.Net.Default,
			Type:      "Service",
			Created:   &c,
			Updated:   &u,
			CreatedBy: service.CreatedBy,
			UpdatedBy: service.CreatedBy,
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
		// Sanitize the scripts
		vpn.Current.PreUp = Sanitize(vpn.Current.PreUp)
		vpn.Current.PostUp = Sanitize(vpn.Current.PostUp)
		vpn.Current.PreDown = Sanitize(vpn.Current.PreDown)
		vpn.Current.PostDown = Sanitize(vpn.Current.PostDown)

		vpn.Current.PersistentKeepalive = 23

		if service.ServiceType != "Ingress" && service.ServiceType != "Egress" {
			vpn.Current.SyncEndpoint = true
		}

		switch service.ServiceType {
		case "Relay":
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)

		case "Tunnel":
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "::/0")

		case "Ingress":
			vpn.Role = "Ingress"
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Default.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "::/0")

		case "Egress":
			vpn.Role = "Egress"
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, vpn.Current.Address...)
			vpn.Current.AllowedIPs = append(vpn.Current.AllowedIPs, "0.0.0.0/0")

		}

		service.VPN, err = CreateVPN(&vpn)
		if err != nil {
			return nil, err
		}
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
	err = mongo.Serialize(service.Id, "id", "services", service)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(service.Id, "id", "services", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	service = v.(*model.Service)

	// return the service
	return service, nil
}

// ReadService service by id
func ReadService(id string) (*model.Service, error) {
	v, err := mongo.Deserialize(id, "id", "services", reflect.TypeOf(model.Service{}))
	if err != nil {
		return nil, err
	}
	service := v.(*model.Service)

	return service, nil
}

// UpdateService preserve keys
func UpdateService(Id string, service *model.Service) (*model.Service, error) {
	v, err := mongo.Deserialize(Id, "id", "services", reflect.TypeOf(model.Service{}))
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

	err = mongo.Serialize(service.Id, "id", "services", service)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "services", reflect.TypeOf(model.Service{}))
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
	v, err := mongo.Deserialize(id, "id", "services", reflect.TypeOf(model.Service{}))
	if err != nil {
		log.Errorf("failed to delete service %s", id)
		return err
	}
	service := v.(*model.Service)

	if service.VPN.Id != "" {
		if service.Server == "" {
			err = DeleteVPN(service.VPN.Id)
			if err != nil {
				log.Errorf("failed to delete vpn %s (%s)", service.VPN.Id, service.VPN.Name)
				return err
			}
		} else {
			// make a device api call to the remote server to delete the vpn
			err = RemoteDelete(vpnAPI, service.Server, service.Device.ApiKey, service.VPN.Id)
			if err != nil {
				log.Errorf("failed to delete remote vpn server %s  key %s vpn %s (%s)", service.Server, service.Device.ApiKey, service.VPN.Id, service.VPN.Name)
			}

		}
	}

	if service.Server == "" && service.Net.Id != "" {
		vpns, err := ReadVPN2("netid", service.Net.Id)
		if err != nil {
			log.Errorf("failed to delete net %s", service.Net.Id)
			return err
		}
		if len(vpns) == 0 {
			err = DeleteNet(service.Net.Id)
			if err != nil {
				log.Errorf("failed to delete net %s (%s)", service.Net.Id, service.Net.NetName)
				return err
			}
		}
	}

	if service.Device.Id != "" {
		if service.Server == "" {
			err = DeleteDevice(service.Device.Id)
			if err != nil {
				log.Errorf("failed to delete device %s (%s)", service.Device.Id, service.Device.Name)
				return err
			}
		} else {
			err = RemoteDelete(deviceAPI, service.Server, service.Device.ApiKey, service.Device.Id)
			if err != nil {
				log.Errorf("failed to delete remote device server %s  key %s device %s (%s)", service.Server, service.Device.ApiKey, service.Device.Id, service.Device.Name)
			}
		}
	}
	// Now delete the service

	err = mongo.Delete(id, "id", "services")
	if err != nil {
		return err
	}

	return nil
}

const (
	deviceAPI = "%s/api/v1.0/device/%s"
	vpnAPI    = "%s/api/v1.0/vpn/%s"
)

// RemoteDelete service
func RemoteDelete(api string, server string, key string, id string) error {
	// make a http DELETE request to the remote server
	var client *http.Client

	if strings.HasPrefix(server, "http:") {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	} else {
		// Create a transport like http.DefaultTransport, but with the configured LocalAddr
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		client = &http.Client{
			Transport: transport,
		}

	}

	var reqURL string = fmt.Sprintf(api, server, id)
	log.Infof("  DELETE %s", reqURL)
	// Create a context with a 15 second timeout
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	if req != nil {
		req.Header.Set("X-API-KEY", key)
		req.Header.Set("User-Agent", "nettica-admin/2.0")
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 401 {
			return fmt.Errorf("Unauthorized")
		} else if resp.StatusCode != 200 {
			log.Errorf("Response Error Code: %v", resp.StatusCode)
			return fmt.Errorf("response error code: %v", resp.StatusCode)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("error reading body %v", err)
			}
			log.Debugf("%s", string(body))
			resp.Body.Close()

			return nil
		}
	} else {
		log.Errorf("ERROR: %v, continuing", err)
	}

	return err
}

func ReadServicesForAccount(accountId string) ([]*model.Service, error) {
	services, err := mongo.ReadServices("accountid", accountId)
	return services, err
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
