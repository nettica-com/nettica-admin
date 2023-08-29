package core

import (
	"errors"
	"os"
	"reflect"
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
		return nil, err
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

	if !fUpdated {
		device.Updated = time.Now().UTC()
	}
	if device.AccountID == "" {
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
	current.ApiID = device.ApiID
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
	current.ClientID = device.ClientID
	current.AuthDomain = device.AuthDomain
	current.AppData = device.AppData
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
	current.Debug = device.Debug
	current.Quiet = device.Quiet
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

// ReadDevices all devices
func ReadDevicesForUser(email string) ([]*model.Device, error) {
	accounts, err := mongo.ReadAllAccounts(email)

	results := make([]*model.Device, 0)

	for _, account := range accounts {
		if account.Status == "Active" {

			if account.NetId != "" {
				// read all the vpns with this netid
				vpns, err := mongo.ReadAllVPNs("netid", account.NetId)
				if err != nil {
					return nil, err
				}
				// read all the devices ...
				devices, err := mongo.ReadAllDevices("accountid", account.Parent)
				if err != nil {
					return nil, err
				}
				// ... and filter them by the vpns
				for _, device := range devices {
					for _, vpn := range vpns {
						if device.Id == vpn.DeviceID {
							device.VPNs = append(device.VPNs, vpn)
							results = append(results, device)
						}
					}
				}
			} else {
				devices, err := mongo.ReadAllDevices("accountid", account.Parent)
				if err != nil {
					return nil, err
				}

				vpns, err := mongo.ReadAllVPNs("accountid", account.Parent)
				if err != nil {
					return nil, err
				}
				for _, device := range devices {
					for _, vpn := range vpns {
						if device.Id == vpn.DeviceID {
							device.VPNs = append(device.VPNs, vpn)
						}
					}
				}
				results = append(results, devices...)
			}
			// If this is a child account, read the child devices
			if account.Id != account.Parent {
				devices, err := mongo.ReadAllDevices("accountid", account.Id)
				if err != nil {
					return nil, err
				}

				vpns, err := mongo.ReadAllVPNs("accountid", account.Id)
				if err != nil {
					return nil, err
				}
				for _, device := range devices {
					for _, vpn := range vpns {
						if device.Id == vpn.DeviceID {
							device.VPNs = append(device.VPNs, vpn)
						}
					}
				}
				results = append(results, devices...)

			}
		}
	}

	return results, err
}
