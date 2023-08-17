package core

import (
	"errors"
	"reflect"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	uuid "github.com/satori/go.uuid"
)

// CreateDevice device with all necessary data
func CreateDevice(device *model.Device) (*model.Device, error) {

	u := uuid.NewV4()
	device.Id = u.String()

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

	// check if device is valid

	err := mongo.Serialize(device.Id, "id", "devices", device)
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

	vpns, err := mongo.ReadAllVPNs("deviceid", device.Id)
	if err != nil {
		return nil, err
	}
	device.Vpns = vpns

	return device, nil
}

// UpdateDevice preserve keys
func UpdateDevice(Id string, device *model.Device, flag bool) (*model.Device, error) {
	v, err := mongo.Deserialize(Id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.Device)

	if current.Id != device.Id {
		return nil, errors.New("records Id mismatch")
	}

	if !flag {
		device.Updated = time.Now().UTC()
	}

	device.Vpns = nil
	err = mongo.Serialize(device.Id, "id", "devices", device)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "devices", reflect.TypeOf(model.Device{}))
	if err != nil {
		return nil, err
	}
	device = v.(*model.Device)
	device.Vpns = current.Vpns

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

				vpns, err := mongo.ReadAllVPNs("netid", account.NetId)
				if err != nil {
					return nil, err
				}
				devices, err := mongo.ReadAllDevices("netid", account.NetId)
				if err != nil {
					return nil, err
				}
				for _, device := range devices {
					for _, vpn := range vpns {
						if device.Id == vpn.DeviceId {
							device.Vpns = append(device.Vpns, vpn)
						}
					}
				}
				results = append(results, devices...)

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
						if device.Id == vpn.DeviceId {
							device.Vpns = append(device.Vpns, vpn)
						}
					}
				}
				results = append(results, devices...)
			}

		}
	}

	return results, err
}
