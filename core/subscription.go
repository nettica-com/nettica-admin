package core

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	template "github.com/nettica-com/nettica-admin/template"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
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

	//	err = SubscriptionEmail(subscription)
	//	if err != nil {
	//		log.Errorf("failed to send email: %v", err)
	//	}

	// data modified, dump new config
	return subscription, nil
}

// EmailHost send email to host
func SubscriptionEmail(sub *model.Subscription) error {
	// get email body
	emailBody, err := template.BillingEmail(*sub)
	if err != nil {
		return err
	}

	// port to int
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), port, os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"))
	s, err := d.Dial()
	if err != nil {
		return err
	}
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetAddressHeader("To", "info@nettica.com", "Nettica")
	m.SetHeader("Subject", "Nettica Apps Billing")
	m.SetBody("text/html", string(emailBody))

	err = gomail.Send(s, m)
	if err != nil {
		return err
	}

	return nil
}

// DeleteSubscription by id
func DeleteSubscription(id string) error {

	v, err := mongo.Deserialize(id, "id", "subscriptions", reflect.TypeOf(model.Subscription{}))
	if err != nil {
		return err
	}

	if v == nil {
		return errors.New("subscription is nil")
	}

	subscription := v.(*model.Subscription)

	errs := subscription.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		return errors.New("failed to validate service")
	}

	lu := time.Now().UTC()
	subscription.LastUpdated = &lu
	subscription.IsDeleted = new(bool)
	*subscription.IsDeleted = true

	err = mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)
	if err != nil {
		return err
	}

	return nil
	/*
		if id == "" {
			return errors.New("id is empty")
		}

		err := mongo.Delete(id, "id", "subscriptions")
		if err != nil {
			return err
		}

		// data modified, dump new config
		return nil
	*/

}

// ReadSubscriptions all clients
func ReadSubscriptions(email string, isDeleted ...bool) ([]*model.Subscription, error) {

	accounts, err := mongo.ReadAllAccounts(email)
	if err != nil {
		return nil, err
	}

	results := make([]*model.Subscription, 0)

	for _, account := range accounts {
		if account.Status == "Active" {
			subscriptions, err := mongo.ReadAllSubscriptions(account.Parent, isDeleted...)
			if err == nil {
				results = append(results, subscriptions...)
			}
		}
	}

	// Update their status
	// this code should no longer be necessary
	// for _, subscription := range results {
	//	if subscription.Status == "active" && subscription.Expires.Before(time.Now().UTC()) && !subscription.Expires.Before(*subscription.Issued) {
	//		subscription.Status = "expired"
	//	}
	// }

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

	offers := model.DiscountOffers{Offers: []string{}}

	if !strings.Contains(account, "account-") {
		return &offers, nil
	}

	deleted, err := mongo.ReadAllSubscriptions(account, true)
	if err != nil {
		return &offers, nil
	}
	if len(deleted) > 0 {
		// if there are deleted subscriptions, we don't offer any discounts
		offers.Offers = []string{}
		return &offers, nil
	}

	subscriptions, err := mongo.ReadAllSubscriptions(account)
	if err != nil {
		return &offers, nil
	}

	if len(subscriptions) == 0 {
		// if there are no subscriptions, we offer the trial discount
		offers.Offers = []string{""}
		return &offers, nil
	}

	/*	for _, subscription := range subscriptions {
			if subscription.Expires.After(time.Now().UTC()) {
				offers.Offers = []string{""}
				break
			}

			if subscription.Expires.Before(time.Now().UTC()) && !subscription.Expires.Before(*subscription.Issued) {
				offers.Offers = []string{"promo"}
			}
		}
	*/

	return &offers, nil
}

func ReadAllTrials() ([]*model.Subscription, error) {
	subscriptions, err := mongo.ReadTrialSubscriptions()
	if err != nil {
		return nil, err
	}

	return subscriptions, nil

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

	if subscription.Status != "cancelled" {
		subscription.Status = "expired"
		subscription.LastUpdated = subscription.Expires
		subscription.UpdatedBy = "system"
	} else {
		last := time.Now().UTC()
		subscription.LastUpdated = &last
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
	isDeleted := false
	subscription.IsDeleted = &isDeleted
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
