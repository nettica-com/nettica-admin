package core

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
)

// ReadServer object, create default one
func ReadServer() ([]*model.Server, error) {

	servers, err := mongo.ReadAllServers()
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// ReadServer2
func ReadServer2(id string) (*model.Server, error) {
	server, err := mongo.Deserialize(id, "serviceGroup", "servers", reflect.TypeOf(model.Server{}))
	if err != nil {
		return nil, err
	}
	return server.(*model.Server), nil
}

// UpdateServer keep private values from existing one
func UpdateServer(server *model.Server) (*model.Server, error) {
	_, err := mongo.Deserialize(server.Id, "id", "servers", reflect.TypeOf(model.Server{}))
	if err != nil {
		return nil, err
	}

	err = mongo.Serialize(server.Id, "id", "servers", server)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// GetAllReservedIps the list of all reserved IPs, client and server
func GetAllReservedIps() ([]string, error) {
	clients, err := ReadHosts()
	if err != nil {
		return nil, err
	}

	reserverIps := make([]string, 0)

	for _, client := range clients {
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

	return reserverIps, nil
}
