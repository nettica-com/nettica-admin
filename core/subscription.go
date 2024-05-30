package core

import (
	"errors"
	"reflect"
	"sort"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	log "github.com/sirupsen/logrus"
)

// CreateSubscription all necessary data
func CreateSubscription(service *model.Subscription) (*model.Subscription, error) {
	/*
		u := uuid.NewV4()
		service.Id = u.String()

		ips := make([]string, 0)
		for _, network := range service.Default.Address {
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

		service.Default.Address = ips
		if len(service.Default.AllowedIPs) == 0 {
			service.Default.AllowedIPs = ips
		}

		service.Created = time.Now().UTC()
		service.Updated = service.Created

		if service.Default.PresharedKey == "" {
			presharedKey, err := wgtypes.GenerateKey()
			if err != nil {
				return nil, err
			}
			service.Default.PresharedKey = presharedKey.String()
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

		err := mongo.Serialize(service.Id, "id", "services", service)
		if err != nil {
			return nil, err
		}

		v, err := mongo.Deserialize(service.Id, "id", "services", reflect.TypeOf(model.Service{}))
		if err != nil {
			return nil, err
		}
		service = v.(*model.Service)

		// data modified, dump new config
		return service, UpdateServerConfigWg()
	*/
	return nil, errors.New("not implemented")
}

// ReadSubscription by id
func ReadSubscription(id string) (*model.Subscription, error) {
	v, err := mongo.Deserialize(id, "id", "subscriptions", reflect.TypeOf(model.Subscription{}))
	if err != nil {
		return nil, err
	}
	subscription := v.(*model.Subscription)

	return subscription, nil
}

// UpdateSubscription by id
func UpdateSubscription(Id string, subscription *model.Subscription) (*model.Subscription, error) {
	v, err := mongo.Deserialize(Id, "id", "subscriptions", reflect.TypeOf(model.Subscription{}))
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New("subscription is nil")
	}

	errs := subscription.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		return nil, errors.New("failed to validate service")
	}

	lu := time.Now().UTC()
	subscription.LastUpdated = &lu

	err = mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "subscriptions", reflect.TypeOf(model.Subscription{}))
	if err != nil {
		return nil, err
	}
	subscription = v.(*model.Subscription)

	// data modified, dump new config
	return subscription, nil
}

// DeleteSubscription by id
func DeleteSubscription(id string) error {

	err := mongo.Delete(id, "id", "subscriptions")
	if err != nil {
		return err
	}

	// data modified, dump new config
	return nil
}

// ReadSubscriptions all clients
func ReadSubscriptions(email string) ([]*model.Subscription, error) {

	accounts, err := mongo.ReadAllAccounts(email)

	results := make([]*model.Subscription, 0)

	for _, account := range accounts {
		if account.Id == account.Parent && account.Status == "Active" {
			subscriptions, err := mongo.ReadAllSubscriptions(account.Email)
			if err == nil {
				results = append(results, subscriptions...)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i] == nil {
			return false
		}
		if results[j] == nil {
			return true
		}
		return results[i].Issued.After(*results[j].Issued)
	})

	return results, err
}
