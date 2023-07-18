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

	if service.RelayHost.MeshId == "" {
		// get all the current meshes and see if there is one with the same name
		meshes, err := ReadMeshes(service.CreatedBy)
		if err != nil {
			return nil, err
		}

		found := false

		for _, m := range meshes {
			if m.MeshName == service.RelayHost.MeshName {
				found = true
				service.RelayHost.MeshName = m.MeshName
				service.RelayHost.MeshId = m.Id
				service.RelayHost.Default = m.Default
				break
			}
		}

		if !found {
			// create a default mesh
			mesh := model.Mesh{
				AccountId:   service.AccountId,
				MeshName:    service.RelayHost.MeshName,
				Description: service.Description,
				Created:     time.Now().UTC(),
				Updated:     time.Now().UTC(),
				CreatedBy:   service.CreatedBy,
			}
			mesh.Default.Address = []string{service.DefaultSubnet}
			mesh.Default.Dns = service.RelayHost.Current.Dns
			mesh.Default.EnableDns = false
			mesh.Default.UPnP = false

			mesh2, err := CreateMesh(&mesh)
			if err != nil {
				return nil, err
			}
			service.RelayHost.MeshName = mesh2.MeshName
			service.RelayHost.MeshId = mesh2.Id
			service.RelayHost.Default = mesh2.Default
		}
	} else {
		// check if mesh exists
		mesh, err := ReadMesh(service.RelayHost.MeshId)
		if err != nil {
			return nil, err
		}
		if mesh == nil {
			return nil, errors.New("mesh does not exist")
		}
		service.RelayHost.MeshName = mesh.MeshName
		service.RelayHost.MeshId = mesh.Id
		service.RelayHost.Default = mesh.Default
		log.Infof("Using existing mesh: %s", mesh.MeshName)
	}

	if service.RelayHost.Id == "" {
		// create a default host using the mesh
		host := model.Host{
			Id:        uuid.NewV4().String(),
			AccountId: service.AccountId,
			Name:      strings.ToLower(service.ServiceType) + "." + service.RelayHost.MeshName,
			Enable:    true,
			MeshId:    service.RelayHost.MeshId,
			MeshName:  service.RelayHost.MeshName,
			HostGroup: service.RelayHost.HostGroup,
			Current:   service.RelayHost.Current,
			Default:   service.RelayHost.Default,
			Type:      "ServiceHost",
			Created:   time.Now().UTC(),
			Updated:   time.Now().UTC(),
			CreatedBy: service.CreatedBy,
		}

		// Failsafe entry for DNS.  Service will break without proper DNS setup.  If nothing is set use google
		if len(host.Current.Dns) == 0 {
			host.Current.Dns = append(host.Current.Dns, "8.8.8.8")
		}

		// Configure the routing for the relay/egress host
		if host.Current.PostUp == "" {
			host.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
		}
		if host.Current.PostDown == "" {
			host.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
		}

		host.Current.PersistentKeepalive = 23

		switch service.ServiceType {
		case "Relay":
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Current.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Default.Address...)

		case "Tunnel":
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Current.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Default.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, "0.0.0.0/0")

		case "Ingress":
			host.Role = "Ingress"
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Current.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Default.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, "0.0.0.0/0")

		case "Egress":
			host.Role = "Egress"
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, host.Current.Address...)
			host.Current.AllowedIPs = append(host.Current.AllowedIPs, "0.0.0.0/0")

		}

		host2, err := CreateHost(&host)
		if err != nil {
			return nil, err
		}
		service.RelayHost = *host2
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

	if service.RelayHost.Id != "" {
		err = DeleteHost(service.RelayHost.Id)
		if err != nil {
			log.Errorf("failed to delete host %s", service.RelayHost.Id)
			return err
		}
	}

	if service.RelayHost.MeshId != "" {
		hosts, err := ReadHost2("meshid", service.RelayHost.MeshId)
		if err != nil {
			log.Errorf("failed to delete mesh %s", service.RelayHost.MeshId)
			return err
		}
		if len(hosts) == 0 {
			err = DeleteMesh(service.RelayHost.MeshId)
			if err != nil {
				log.Errorf("failed to delete mesh %s", service.RelayHost.MeshId)
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

// ReadServiceHost returns all services configured for a host
func ReadServiceHost(serviceGroup string) ([]*model.Service, error) {
	services, err := mongo.ReadServiceHost(serviceGroup)
	return services, err
}
