package core

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
)

// CreateDevice device with all necessary data
func CreateDevice(device *model.Device) (*model.Device, error) {

	var err error
	device.Id, err = util.RandomString(12)
	if err != nil {
		return nil, err
	}
	device.Id = "device-" + device.Id

	device.Created = time.Now().UTC()
	device.Updated = device.Created

	if device.ApiKey == "" {
		var err error
		device.ApiKey, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
		device.ApiKey = "device-api-" + device.ApiKey
	}

	if device.Server == "" {
		device.Server = os.Getenv("SERVER")
	}

	// check if device is valid
	errs := device.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("device validation error")
		}
		return nil, errors.New("failed to validate device")
	}

	err = mongo.Serialize(device.Id, "id", "devices", device)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(device.Id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	device = v.(*model.Device)

	// data modified, dump new config
	return device, nil
}

// ReadDevice device by id
func ReadDevice(id string) (*model.Device, error) {
	v, err := mongo.Deserialize(id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		if strings.HasPrefix(id, "ez-") {

			v, err = mongo.Deserialize(id, "ezcode", "devices", reflect.TypeOf(model.Device{}))
			if err != nil {
				return nil, err
			}
		} else {

			id = strings.TrimPrefix(id, "device-id-")

			if id == "" || strings.HasPrefix(id, "device-") {
				return nil, err
			}

			v, err = mongo.Deserialize(id, "instanceid", "devices", reflect.TypeOf(model.Device{}))
			if err != nil {
				return nil, err
			}
		}
	}
	device := v.(*model.Device)

	//	vpns, err := mongo.ReadAllVPNs("deviceid", device.Id)
	//	if err != nil {
	//		return nil, err
	//	}
	//	device.VPNs = vpns

	return device, nil
}

// UpdateDevice preserve keys
func UpdateDevice(Id string, device *model.Device, fUpdated bool) (*model.Device, error) {
	v, err := mongo.Deserialize(Id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.Device)

	if current.Id != device.Id {
		return nil, errors.New("records Id mismatch")
	}

	if current.Updated.After(device.Updated) {
		log.Errorf("UpdateDevice: device %s has been updated by another user", device.Id)
	}

	if !fUpdated {
		device.Updated = time.Now().UTC()
	}

	if !strings.HasPrefix(device.AccountID, "account-") {
		device.AccountID = current.AccountID
	}

	device.VPNs = nil

	// check if device is valid
	errs := device.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("device validation error")
		}
		return nil, errors.New("failed to validate device")
	}

	// copy each field individually
	if device.Name != "" {
		current.Name = device.Name
	}
	current.Description = device.Description
	if device.AccountID != "" {
		current.AccountID = device.AccountID
	}
	current.Updated = device.Updated
	if device.Server != "" {
		current.Server = device.Server
	}
	if device.ApiKey != "" {
		current.ApiKey = device.ApiKey
	}
	if device.UpdatedBy != "" {
		current.UpdatedBy = device.UpdatedBy
	} else {
		if !fUpdated {
			current.UpdatedBy = "API"
		}
	}
	current.Enable = device.Enable
	current.Tags = device.Tags
	current.Platform = device.Platform
	current.OS = device.OS
	if current.Platform == "" {
		if current.OS == "windows" {
			current.Platform = "Windows"
		}
		if current.OS == "linux" {
			current.Platform = "Linux"
		}
	}
	current.Architecture = device.Architecture
	current.Version = device.Version
	if device.CheckInterval != 0 {
		current.CheckInterval = device.CheckInterval
	}
	if current.CheckInterval == 0 {
		current.CheckInterval = 10
	}
	if device.ServiceGroup != "" {
		current.ServiceGroup = device.ServiceGroup
	}
	if device.ServiceApiKey != "" {
		current.ServiceApiKey = device.ServiceApiKey
	}

	// stupidity is not a right
	changed := false
	if current.Registered != device.Registered {
		changed = true
	}
	current.Registered = device.Registered
	if (current.InstanceID != device.InstanceID && !changed) ||
		(current.EZCode != device.EZCode && !changed) {
		current.Registered = false
	}
	current.InstanceID = device.InstanceID
	current.EZCode = device.EZCode
	current.LastSeen = device.LastSeen

	err = mongo.Serialize(device.Id, "id", "devices", current)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	device = v.(*model.Device)
	device.VPNs = current.VPNs

	// data modified, dump new config
	return device, nil
}

// DeleteDevice from database
func DeleteDevice(id string) error {

	vpns, err := mongo.ReadAllVPNs("deviceid", id)

	if err != nil {
		return err
	}
	for _, vpn := range vpns {
		err = DeleteVPN(vpn.Id)
		if err != nil {
			return err
		}
	}

	return mongo.Delete(id, "id", "devices")
}

// ReadDeviceByApiKey(device.ApiKey)
func ReadDeviceByApiKey(apikey string) (*model.Device, error) {
	v, err := mongo.Deserialize(apikey, "apiKey", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	device := v.(*model.Device)

	return device, nil
}

// ReadDevice2 device by param and id
func ReadDevice2(param string, id string) ([]*model.Device, error) {
	return mongo.ReadAllDevices(param, id)
}

// ReadDevices all devices
func ReadDevices() ([]*model.Device, error) {
	return mongo.ReadAllDevices("", "")
}

// ReadDevicesForAccount
func ReadDevicesForAccount(accountid string) ([]*model.Device, error) {
	return mongo.ReadAllDevices("accountid", accountid)
}

// ReadDevices all devices
// This code needs a severe rewrite
func ReadDevicesForUser(email string) ([]*model.Device, error) {
	accounts, err := mongo.ReadAllAccounts(email)
	if err != nil {
		return nil, err
	}

	results := make([]*model.Device, 0)

	for _, account := range accounts {
		if account.Status == "Active" {
			if account.Role == "User" || account.Role == "Guest" {
				// users and guests cannot see devices they did not create
			} else {

				if account.NetId != "" {

					// read all the vpns with this netid
					vpns, err := mongo.ReadVPNsforNetwork(account.NetId)
					if err != nil {
						return nil, err
					}

					for _, vpn := range vpns {
						device := vpn.Devices[0]
						vpn.Devices = nil
						device.VPNs = append(device.VPNs, vpn)
						results = append(results, device)
					}
				} else {

					devices, err := mongo.ReadDevicesAndVPNsForAccount(account.Parent)
					if err != nil {
						return nil, err
					}
					results = append(results, devices...)
				}
			}
		}
	}

	//
	// This needs to be reworked as it doesn't take into account
	// the hidden status of the user
	//

	// now handle users and guests who can only see devices they created
	vpns, err := mongo.ReadAllVPNs("createdBy", email)
	if err != nil {
		return nil, err
	}

	// now read devices created by this user and add any missing
	devices, err := mongo.ReadAllDevices("createdBy", email)
	if err != nil {
		return nil, err
	}

	// loop through the results and add any missing devices
	for _, device := range devices {

		// add the device if it hasn't already been added
		found := false
		for _, result := range results {
			if device.Id == result.Id {
				found = true
				break
			}
		}
		if !found {
			// associate any vpns to the device
			for _, vpn := range vpns {
				if device.Id == vpn.DeviceID {
					device.VPNs = append(device.VPNs, vpn)
				}
			}

			results = append(results, device)
		}
	}

	return results, err
}
