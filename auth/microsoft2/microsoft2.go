package microsoft2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	mongodb "github.com/nettica-com/nettica-admin/mongo"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	//	"gopkg.in/auth0.v4"
	"os"
)

// Oauth2Microsoft in order to implement interface, struct is required
type Oauth2Microsoft struct{}

var (
	oauth2Config *oauth2.Config
	userCache    *cache.Cache
	clientApp    confidential.Client
	publicApp    public.Client
)

// Setup validate provider
func (o *Oauth2Microsoft) Setup() error {
	var err error

	userCache = cache.New(60*time.Minute, 10*time.Minute)

	cred, err := confidential.NewCredFromSecret(os.Getenv("OAUTH2_CLIENT_SECRET"))
	if err != nil {
		return err
	}

	// 	oidcProvider, err = oidc.NewProvider(context.TODO(), os.Getenv("OAUTH2_PROVIDER"))

	// Create a confidential client using a client ID and secret
	clientApp, err = confidential.New(os.Getenv("OAUTH2_PROVIDER"), os.Getenv("OAUTH2_CLIENT_ID"), cred)
	if err != nil {
		return err
	}

	publicApp, err = public.New(os.Getenv("OAUTH2_CLIENT_ID"))
	if err != nil {
		return err
	}

	endpoint := oauth2.Endpoint{
		AuthURL:   os.Getenv("OAUTH2_PROVIDER") + "/authorize",
		TokenURL:  os.Getenv("OAUTH2_PROVIDER") + "/token",
		AuthStyle: oauth2.AuthStyleAutoDetect}

	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     endpoint,
	}

	return nil
}

// CodeUrl get url to redirect client for auth
func (o *Oauth2Microsoft) CodeUrl(state string) string {

	// authURL, err := app.AuthCodeURL(context.Background(), confidentialConfig.ClientID, confidentialConfig.RedirectURI, confidentialConfig.Scopes)

	authUrl, err := clientApp.AuthCodeURL(context.Background(), oauth2Config.ClientID, oauth2Config.RedirectURL, oauth2Config.Scopes)
	if err != nil {
		log.Error(err)
	}

	authUrl = authUrl + "&state=" + state
	log.Infof("authUrl: %s", authUrl)

	return authUrl
}

func (o *Oauth2Microsoft) CodeUrl2(state string) string {

	client_id := os.Getenv("OAUTH2_AGENT_CLIENT_ID")
	redirect_uri := os.Getenv("OAUTH2_AGENT_REDIRECT_URL")

	authUrl, err := publicApp.AuthCodeURL(context.Background(), client_id, redirect_uri, oauth2Config.Scopes)
	if err != nil {
		log.Error(err)
	}

	authUrl = authUrl + "&state=" + state
	log.Infof("authUrl: %s", authUrl)

	return authUrl
}

// Exchange exchange code for Oauth2 token
func (o *Oauth2Microsoft) Exchange(code string) (*oauth2.Token, error) {
	// oauth2Token, err := oauth2Config.Exchange(context.TODO(), code)

	authResult, err := clientApp.AcquireTokenByAuthCode(context.Background(), code, oauth2Config.RedirectURL, oauth2Config.Scopes)
	if err != nil {
		return nil, err
	}

	// create an oauth2.Token from the AuthResult and including the IDToken
	//oauth2Token := oauth2.Token{

	oauth2Token := &oauth2.Token{
		AccessToken: authResult.AccessToken,
		Expiry:      authResult.ExpiresOn,
	}
	oauth2Token = oauth2Token.WithExtra(map[string]interface{}{ // Add the ID token to the extra parameters
		"id_token":    authResult.IDToken,
		"auth_result": authResult})

	return oauth2Token, nil
}

func (o *Oauth2Microsoft) Exchange2(code string) (*oauth2.Token, error) {
	authResult, err := publicApp.AcquireTokenByAuthCode(context.Background(), code, "com.nettica.agent://callback/agent", oauth2Config.Scopes)
	if err != nil {
		return nil, err
	}

	log.Infof("authResult: %v", authResult)

	// create an oauth2.Token from the AuthResult and including the IDToken
	//oauth2Token := oauth2.Token{

	oauth2Token := &oauth2.Token{
		AccessToken: authResult.AccessToken,
		Expiry:      authResult.ExpiresOn,
	}
	oauth2Token = oauth2Token.WithExtra(map[string]interface{}{ // Add the ID token to the extra parameters
		"id_token":    authResult.IDToken,
		"auth_result": authResult})

	return oauth2Token, nil
}

// UserInfo get token user
func (o *Oauth2Microsoft) UserInfo(oauth2Token *oauth2.Token) (*model.User, error) {

	var err error

	authResult, ok := oauth2Token.Extra("auth_result").(confidential.AuthResult)
	if !ok {
		return nil, fmt.Errorf("no auth_result field in oauth2 token")
	}

	cacheUser, _ := userCache.Get(oauth2Token.AccessToken)
	if cacheUser != nil {
		return cacheUser.(*model.User), nil
	}

	// get some infos about user
	user := &model.User{}
	user.Sub = authResult.IDToken.Subject
	user.Email = authResult.IDToken.Email
	user.Email = strings.ToLower(user.Email)

	log.Infof("user.Sub: %s", user.Sub)

	user.Name = authResult.IDToken.Name
	user.Picture = os.Getenv("SERVER") + "/account-circle.png"

	user.Issuer = authResult.IDToken.Issuer
	user.IssuedAt = time.Unix(authResult.IDToken.IssuedAt, 0)
	log.Infof("user %s token expires %v", user.Email, authResult.IDToken.ExpirationTime)

	accounts, err := mongodb.ReadAllAccounts(user.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = "Me"
			account.Sub = user.Sub
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
