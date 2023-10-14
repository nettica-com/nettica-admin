package core

import (
	"os"
	"reflect"

	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"
)

func EnforceLimits() bool {

	if (os.Getenv("ENFORCE_LIMITS") == "1") || (os.Getenv("ENFORCE_LIMITS") == "true") {
		return true
	}

	return false
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
