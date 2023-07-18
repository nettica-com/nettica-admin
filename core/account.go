package core

import (
	"errors"
	"reflect"
	"strings"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
)

// CreateAccount with all necessary data
func CreateAccount(account *model.Account) (*model.Account, error) {

	var err error

	if account.Id == "" {
		account.Id, err = util.RandomString(16)
		if err != nil {
			return nil, err
		}
	}

	if account.Key == "" {
		account.Key, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
	}

	if account.Parent == "" {
		account.Parent = account.Id
	}

	account.Created = time.Now()

	errs := account.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("account validation error")
		}
		return nil, errors.New("failed to validate account")
	}

	err = mongo.Serialize(account.Id, "id", "accounts", account)

	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(account.Id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}
	account = v.(*model.Account)

	// return current account
	return account, nil
}

// ReadACcount by id
func ReadAccount(id string) (*model.Account, error) {

	v, err := mongo.Deserialize(id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}
	account := v.(*model.Account)

	return account, nil
}

// ReadAllAccounts account by id or email address
func ReadAllAccounts(email string) ([]*model.Account, error) {

	if strings.Contains(email, "@") {
		return mongo.ReadAllAccounts(email)
	} else {
		return mongo.ReadAllAccountsForID(email)
	}
}

// UpdateUser preserve keys
func UpdateAccount(Id string, user *model.Account) (*model.Account, error) {
	v, err := mongo.Deserialize(Id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.Account)

	if current != nil && user != nil &&
		current.Email != user.Email {
		return nil, errors.New("records Id mismatch")
	}

	// check if user is valid
	errs := user.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("user validation error")
		}
		return nil, errors.New("failed to validate user")
	}

	err = mongo.Serialize(Id, "id", "accounts", user)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}
	user = v.(*model.Account)

	// data modified, dump new config
	return user, nil
}

// DeleteAccount from mongo
func DeleteAccount(id string) error {

	return mongo.Delete(id, "id", "accounts")
}

// ActivateAccount when joining
func ActivateAccount(id string) (string, error) {

	var a *model.Account

	v, err := mongo.Deserialize(id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return "Error", err
	}
	a = v.(*model.Account)
	a.Status = "Active"

	err = mongo.Serialize(id, "id", "accounts", a)
	if err != nil {
		return "Error", err
	}

	log.Infof("Account Activated: %s %s", a.Email, id)

	return "Account activated.", nil
}
