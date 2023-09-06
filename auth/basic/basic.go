package basic

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"encoding/base64"
	"encoding/json"

	mongodb "github.com/nettica-com/nettica-admin/mongo"
	log "github.com/sirupsen/logrus"

	"github.com/coreos/go-oidc"

	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/shadow"
	util "github.com/nettica-com/nettica-admin/util"
	"golang.org/x/oauth2"
)

type Oauth2Basic struct{}

// Create an OAuth2 shim for BasicAuth
func (o *Oauth2Basic) Setup() error {
	return nil
}

func (o *Oauth2Basic) Logout() error {
	return nil
}

// CodeUrl get url to redirect client for auth
func (o *Oauth2Basic) CodeUrl(state string) string {
	return "/login"
}

// Exchange exchange code for Oauth2 token
func (o *Oauth2Basic) Exchange(code string) (*oauth2.Token, error) {

	// code contains the username and password base64 encoded
	// base64 decode the string
	userpass, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return nil, err
	}

	// split the username and password
	parts := strings.SplitN(string(userpass), ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid username and password")
	}
	user := parts[0]
	pass := parts[1]

	// validate the username and password
	err = shadow.ShadowAuthPlain(user, pass)
	if err != nil {
		return nil, err
	}

	rand, err := util.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  rand,
		TokenType:    "Bearer",
		RefreshToken: "",
		Expiry:       time.Now().Add(time.Hour * 24),
	}
	// add the user to the token
	idtoken := &oidc.IDToken{Subject: user, Issuer: "Basic", IssuedAt: time.Now(), Expiry: time.Now().Add(time.Hour * 24)}

	m := make(map[string]interface{})

	raw, err := json.Marshal(idtoken)
	if err != nil {
		return nil, err
	}
	m["id_token"] = string(raw)
	token = token.WithExtra(m)

	return token, nil
}

// UserInfo get token user
func (o *Oauth2Basic) UserInfo(oauth2Token *oauth2.Token) (*model.User, error) {
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token field in oauth2 token")
	}

	verified := false
	var idToken *oidc.IDToken
	var err error

	// decode the json string into an idToken
	err = json.Unmarshal([]byte(rawIDToken), &idToken)
	if err != nil {
		return nil, err
	}

	verified = true

	if !verified || err != nil {
		return nil, err
	}

	// get some infos about user
	user := &model.User{}
	user.Sub = idToken.Subject
	user.Email = idToken.Subject + "@localhost"
	user.Email = strings.ToLower(user.Email)
	user.Picture = "/account-circle.svg"
	user.Issuer = idToken.Issuer
	user.IssuedAt = idToken.IssuedAt
	log.Infof("user %v", user)

	// check if user exists
	accounts, err := mongodb.ReadAllAccounts(user.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			host, _ := os.Hostname()
			account.AccountName = host
			account.Name = user.Sub
			account.Email = user.Email
			account.Role = "Owner"
			account.Status = "Active"
			account.CreatedBy = user.Email
			account.UpdatedBy = user.Email
			account.Picture = user.Picture
			a, err := core.CreateAccount(&account)
			log.Infof("CREATE ACCOUNT = %v", a)
			if err != nil {
				log.Error(err)
			}
			accounts, err = mongodb.ReadAllAccounts(user.Email)
			if err != nil {
				log.Error(err)
			}

		}
	}
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Id == accounts[i].Parent {
			user.AccountID = accounts[i].Id
			break
		}
	}
	if user.AccountID == "" {
		user.AccountID = accounts[0].Id
	}

	err = mongodb.UpsertUser(user)
	if err != nil {
		log.Error(err)
	}
	return user, nil
}
