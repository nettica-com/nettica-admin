package core

import (
	"os"
	"reflect"
	"strconv"

	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"
)

func EnforceLimits() bool {

	if (os.Getenv("ENFORCE_LIMITS") == "1") || (os.Getenv("ENFORCE_LIMITS") == "true") {
		return true
	}

	return false
}

func GetDefaultMaxMembers() int {
	if os.Getenv("LIMITS_DEFAULT_MAX_MEMBERS") != "" {
		x, err := strconv.ParseInt(os.Getenv("DEFAULT_MAX_MEMBERS"), 10, 32)
		if err == nil {
			return int(x)
		}
	}
	return 3
}

func GetDefaultMaxNetworks() int {
	if os.Getenv("LIMITS_DEFAULT_MAX_NETWORKS") != "" {
		x, err := strconv.ParseInt(os.Getenv("DEFAULT_MAX_NETWORKS"), 10, 32)
		if err == nil {
			return int(x)
		}
	}
	return 2
}

func GetDefaultMaxDevices() int {
	if os.Getenv("LIMITS_DEFAULT_MAX_DEVICES") != "" {
		x, err := strconv.ParseInt(os.Getenv("DEFAULT_MAX_DEVICES"), 10, 32)
		if err == nil {
			return int(x)
		}
	}
	return 5
}

func GetDefaultMaxServices() int {
	if os.Getenv("LIMITS_DEFAULT_MAX_SERVICES") != "" {
		x, err := strconv.ParseInt(os.Getenv("DEFAULT_MAX_SERVICES"), 10, 32)
		if err == nil {
			return int(x)
		}
	}
	return 0
}

func GetDefaultTolerance() float64 {
	if os.Getenv("LIMITS_DEFAULT_TOLERANCE") != "" {
		x, err := strconv.ParseFloat(os.Getenv("DEFAULT_TOLERANCE"), 64)
		if err == nil {
			return x
		}
	}
	return 1.0
}

func ReadLimits(accountid string) (*model.Limits, error) {
	var limit *model.Limits

	v, err := mongo.Deserialize(accountid, "accountid", "limits", reflect.TypeOf(model.Limits{}))
	if err != nil {
		return nil, err
	}

	limit = v.(*model.Limits)

	return limit, err

}
