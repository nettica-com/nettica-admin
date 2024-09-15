package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	mongodb "github.com/nettica-com/nettica-admin/mongo"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"os"
)

// OAuth2Google in order to implement interface, struct is required
type OAuth2Google struct{}

var (
	oauth2Config        *oauth2.Config
	oidcProvider        *oidc.Provider
	oidcIDTokenVerifier []*oidc.IDTokenVerifier
	userCache           *cache.Cache
)

// Setup validate provider
func (o *OAuth2Google) Setup() error {
	var err error

	userCache = cache.New(60*time.Minute, 10*time.Minute)
	oidcProvider, err = oidc.NewProvider(context.TODO(), os.Getenv("OAUTH2_PROVIDER"))
	if err != nil {
		return err
	}

	oidcIDTokenVerifier = make([]*oidc.IDTokenVerifier, 0)
	oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID")}))
	//oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID_WINDOWS")}))

	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     oidcProvider.Endpoint(),
	}

	return nil
}

// CodeUrl get url to redirect client for auth
func (o *OAuth2Google) CodeUrl(state string) string {
	return oauth2Config.AuthCodeURL(state)
}

func (o *OAuth2Google) CodeUrl2(state string) string {
	return o.CodeUrl(state)
}

// Exchange exchange code for Oauth2 token
func (o *OAuth2Google) Exchange(code string) (*oauth2.Token, error) {
	oauth2Token, err := oauth2Config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}

func (o *OAuth2Google) Exchange2(code string) (*oauth2.Token, error) {
	token, err := o.Exchange(code)
	return token, err
}

// UserInfo get token user
func (o *OAuth2Google) UserInfo(oauth2Token *oauth2.Token) (*model.User, error) {
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token field in oauth2 token")
	}

	verified := false
	var idToken *oidc.IDToken
	var err error

	for _, verifier := range oidcIDTokenVerifier {
		idToken, err = verifier.Verify(context.TODO(), rawIDToken)
		if err == nil {
			verified = true
			break
		}
	}

	if !verified || err != nil {
		return nil, err
	}

	cacheUser, _ := userCache.Get(oauth2Token.AccessToken)
	if cacheUser != nil {
		return cacheUser.(*model.User), nil
	}

	// ID Token payload is just JSON
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to get id token claims: %s", err)
	}

	// get some infos about user
	user := &model.User{}
	user.Sub = claims["sub"].(string)
	user.Email = claims["email"].(string)
	user.Email = strings.ToLower(user.Email)

	log.Infof("user.Sub: %s", user.Sub)

	if v, found := claims["name"]; found && v != nil {
		user.Name = v.(string)
	} else {
		log.Error("name not found in user info claims")
	}

	if v, found := claims["picture"]; found && v != nil {
		user.Picture = v.(string)
	} else {
		user.Picture = "/account-circle.svg"
	}

	user.Issuer = idToken.Issuer
	user.IssuedAt = idToken.IssuedAt
	log.Infof("user %s token expires %v", user.Email, idToken.Expiry)

	accounts, err := mongodb.ReadAllAccounts(user.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = "Me"
			account.AccountName = "Company"
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
	userCache.Set(oauth2Token.AccessToken, user, 0)
	return user, nil
}
