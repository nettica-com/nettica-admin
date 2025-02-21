package core

import (
	"errors"
	"reflect"
	"sort"
	"strings"
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

	if id == "" {
		return errors.New("id is empty")
	}

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
		if account.Status == "Active" {
			subscriptions, err := mongo.ReadAllSubscriptions(account.Parent)
			if err == nil {
				results = append(results, subscriptions...)
			}
		}
	}

	// Update their status
	for _, subscription := range results {
		if subscription.Expires.After(time.Now().UTC()) {
			subscription.Status = "active"
		} else if subscription.Expires.Before(time.Now().UTC()) && !subscription.Expires.Before(*subscription.Issued) {
			subscription.Status = "expired"
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

func GetOffers(account string) (*model.DiscountOffers, error) {

	offers := model.DiscountOffers{Offers: []string{"intro"}}

	if !strings.Contains(account, "account-") {
		return &offers, nil
	}

	subscriptions, err := mongo.ReadAllSubscriptions(account)
	if err != nil {
		return &offers, nil
	}

	for _, subscription := range subscriptions {
		if subscription.Expires.After(time.Now().UTC()) {
			offers.Offers = []string{""}
			break
		}

		if subscription.Expires.Before(time.Now().UTC()) && !subscription.Expires.Before(*subscription.Issued) {
			offers.Offers = []string{"promo"}
		}
	}

	return &offers, nil
}

// ExpireSubscription by id
func ExpireSubscription(id string) error {

	subscription, err := ReadSubscription(id)
	if err != nil {
		return err
	}

	if subscription.Status == "expired" {
		return errors.New("subscription is expired")
	}

	if subscription.Expires.After(time.Now().UTC()) {
		return errors.New("subscription has not expired")
	}

	if subscription.Status == "active" {
		subscription.Status = "expired"
		subscription.LastUpdated = subscription.Expires
		subscription.UpdatedBy = "system"
	} else {
		last := time.Now().UTC()
		subscription.LastUpdated = &last
		subscription.UpdatedBy = "system"
	}
	_, err = UpdateSubscription(subscription.Id, subscription)
	if err != nil {
		log.Errorf("failed to update subscription: %v", err)
	}

	// get the total number of credits available
	subscriptions, err := ReadSubscriptions(subscription.AccountID)
	if err != nil {
		return err
	}

	total_credits := 0
	for _, s := range subscriptions {
		if s.Status == "active" {
			total_credits += s.Credits
		}
	}

	// get all the active services running on this subscription
	services, err := mongo.ReadAllServices(subscription.AccountID)
	if err != nil {
		return err
	}

	available := total_credits - subscription.Credits

	if available >= 0 {
		// nothing to expire, they still have other active subscriptions
		return nil
	}

	// expire some or all services

	// at this pont available is negative
	// we expire services until we reach 0 or all services are expired
	for _, service := range services {

		// only disable active services
		if service.Device.Enable {
			service.Device.Enable = false

			device, err := ReadDevice(service.Device.Id)
			if err == nil {
				device.Enable = false
				_, err = UpdateDevice(device.Id, device, false)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("failed to update device")
				}
			}

			available++

			if available >= 0 {
				break
			}
		}
	}

	return nil
}

func CancelSubscription(id string) error {

	subscription, err := ReadSubscription(id)
	if err != nil {
		return err
	}

	if subscription.Status == "cancelled" {
		return errors.New("subscription is already cancelled")
	}

	subscription.Status = "cancelled"
	last := time.Now().UTC()
	subscription.LastUpdated = &last
	_, err = UpdateSubscription(subscription.Id, subscription)
	if err != nil {
		return err
	}

	err = ExpireSubscription(subscription.Id)

	return err
}

// RenewSubscription by id
func RenewSubscription(id string) error {

	subscription, err := ReadSubscription(id)
	if err != nil {
		return err
	}

	if subscription.Status == "active" {
		return nil
	}

	last := time.Now().UTC()
	subscription.Status = "active"
	subscription.LastUpdated = &last
	_, err = UpdateSubscription(subscription.Id, subscription)
	if err != nil {
		return err
	}

	// reactivate some services
	subscriptions, err := ReadSubscriptions(subscription.AccountID)
	if err != nil {
		return err
	}

	total_credits := 0
	for _, s := range subscriptions {
		if s.Status == "active" {
			total_credits += s.Credits
		}
	}

	// get all the active services running on this subscription
	services, err := mongo.ReadAllServices(subscription.AccountID)
	if err != nil {
		return err
	}

	for _, service := range services {

		// only enable inactive services
		if !service.Device.Enable {
			service.Device.Enable = true

			device, err := ReadDevice(service.Device.Id)
			if err == nil {
				device.Enable = true
				_, err = UpdateDevice(device.Id, device, false)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("failed to update device")
				}
			}

			total_credits--

			if total_credits <= 0 {
				break
			}
		}
	}

	return nil
}
func GetSubscriptionByReceipt(receipt string) (*model.Subscription, error) {

	v, err := mongo.Deserialize(receipt, "receipt", "subscriptions", reflect.TypeOf(model.Subscription{}))
	if err != nil {
		return nil, err
	}

	subscription := v.(*model.Subscription)

	return subscription, nil
}
