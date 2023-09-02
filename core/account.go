package core

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	template "github.com/nettica-com/nettica-admin/template"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// CreateAccount with all necessary data
func CreateAccount(account *model.Account) (*model.Account, error) {

	var err error

	if account.Id == "" {
		account.Id, err = util.RandomString(16)
		if err != nil {
			return nil, err
		}
		account.Id = "account-" + account.Id
	}

	if account.ApiKey == "" {
		account.ApiKey, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
		account.ApiKey = "nettica-api-" + account.ApiKey
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

func GetAccount(email string, accountid string) (*model.Account, error) {

	account, err := mongo.ReadAccountForUser(email, accountid)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func GetAccountFromApiKey(apikey string) (*model.Account, error) {

	v, err := mongo.Deserialize(apikey, "apiKey", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}

	account := v.(*model.Account)

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
func ActivateAccount(id string) (*model.Account, error) {

	var a *model.Account

	v, err := mongo.Deserialize(id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return nil, err
	}
	a = v.(*model.Account)
	if a.Status != "Suspended" {
		a.Status = "Active"
	} else {
		return nil, errors.New("account is suspended")
	}

	err = mongo.Serialize(id, "id", "accounts", a)
	if err != nil {
		return nil, err
	}
	if a.NetName == "" {
		a.NetName = "All Networks"
	}

	log.Infof("Account Activated: %s %s", a.Email, id)

	return a, nil
}

func Email(id string) error {
	v, err := mongo.Deserialize(id, "id", "accounts", reflect.TypeOf(model.Account{}))
	if err != nil {
		return err
	}

	account := v.(*model.Account)
	net := account.NetName

	if net == "" {
		net = "All Networks"
	}

	// get email body
	emailBody, err := template.DumpUserEmail(id, net)
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
	m.SetAddressHeader("To", id, id)
	m.SetHeader("Subject", "nettica.com Invitation")
	m.SetBody("text/html", string(emailBody))

	err = gomail.Send(s, m)
	if err != nil {
		return err
	}

	return nil
}
