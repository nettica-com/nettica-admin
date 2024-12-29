package oauth2oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	mongodb "github.com/nettica-com/nettica-admin/mongo"
	"github.com/nettica-com/nettica-admin/util"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	//	"gopkg.in/auth0.v4"
	"os"
)

// Oauth2idc in order to implement interface, struct is required
type Oauth2idc struct{}

var (
	oauth2Config        *oauth2.Config
	oidcProvider        *oidc.Provider
	oidcIDTokenVerifier []*oidc.IDTokenVerifier
	userCache           *cache.Cache

	publicConfig   *oauth2.Config
	publicProvider *oidc.Provider
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
}

// Setup validate provider
func (o *Oauth2idc) Setup() error {
	var err error

	userCache = cache.New(60*time.Minute, 10*time.Minute)

	oidcProvider, err = oidc.NewProvider(context.TODO(), os.Getenv("OAUTH2_PROVIDER"))
	if err != nil {
		return err
	}

	publicProvider, err = oidc.NewProvider(context.TODO(), os.Getenv("OAUTH2_AGENT_PROVIDER"))
	if err != nil {
		return err
	}

	oidcIDTokenVerifier = make([]*oidc.IDTokenVerifier, 0)
	oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID")}))
	oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_AGENT_CLIENT_ID")}))

	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     oidcProvider.Endpoint(),
	}

	publicConfig = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_AGENT_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_AGENT_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_AGENT_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     publicProvider.Endpoint(),
	}

	return nil
}

// CodeUrl get url to redirect client for auth
func (o *Oauth2idc) CodeUrl(state string) string {
	return oauth2Config.AuthCodeURL(state)
}

func (o *Oauth2idc) CodeUrl2(state string) string {

	client_id := os.Getenv("OAUTH2_AGENT_CLIENT_ID")
	redirect_url := os.Getenv("OAUTH2_AGENT_REDIRECT_URL")
	audience := os.Getenv("OAUTH2_AGENT_AUDIENCE")
	provider := os.Getenv("OAUTH2_AGENT_PROVIDER")

	url := provider + "authorize?response_type=code&client_id=" + client_id + "&redirect_uri=" + redirect_url + "&audience=" + audience + "&state=" + state + "&scope=openid%20profile%20email"

	return url

}

// Exchange exchange code for Oauth2 token
func (o *Oauth2idc) Exchange(code string) (*oauth2.Token, error) {
	oauth2Token, err := oauth2Config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}
func (o *Oauth2idc) Exchange2(code string) (*oauth2.Token, error) {

	// Make a http request using the agent configuration information
	client_id := os.Getenv("OAUTH2_AGENT_CLIENT_ID")
	client_secret := os.Getenv("OAUTH2_AGENT_CLIENT_SECRET")
	redirect_url := os.Getenv("OAUTH2_AGENT_REDIRECT_URL")
	audience := os.Getenv("OAUTH2_AGENT_AUDIENCE")

	provider := os.Getenv("OAUTH2_AGENT_PROVIDER")

	// make an http post to the oauth2 token endpoint
	// with the code and other required parameters
	// to get the access token
	// and other information

	httpClient := &http.Client{}
	rsp, err := httpClient.PostForm(provider+"oauth/token/", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {client_id},
		"client_secret": {client_secret},
		"redirect_uri":  {redirect_url},
		"code":          {code},
		"audience":      {audience},
	})
	if err != nil {
		log.Info(err)
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	log.Infof("body: %s", body)

	// read the response body and serialize it into a TokenResponse struct
	var tokenResponse TokenResponse

	// decode the body bytes into the struct

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	oauth2Token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		Expiry:      time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
		TokenType:   tokenResponse.TokenType,
	}
	oauth2Token = oauth2Token.WithExtra(map[string]interface{}{ // Add the ID token to the extra parameters
		"id_token": tokenResponse.IdToken})

	//	oauth2Token, err := publicConfig.Exchange(context.TODO(), code)
	//	if err != nil {
	//		return nil, err
	//	}

	return oauth2Token, nil
}

// UserInfo get token user
func (o *Oauth2idc) UserInfo(oauth2Token *oauth2.Token) (*model.User, error) {
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

	userInfo, err := oidcProvider.UserInfo(context.TODO(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		return nil, err
	}

	// ID Token payload is just JSON
	var claims map[string]interface{}
	if err := userInfo.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to get id token claims: %s", err)
	}

	// get some infos about user
	user := &model.User{}
	user.Sub = userInfo.Subject
	user.Email = strings.ToLower(userInfo.Email)
	user.Profile = userInfo.Profile

	//	for k, v :=  range claims {
	//		user.Claims = user.Claims + "<br>" + k + ":" + fmt.Sprintf("%v", v)
	//	}

	log.Infof("user.Sub: %s", user.Sub)

	if v, found := claims["name"]; found && v != nil {
		user.Name = v.(string)
	} else {
		log.Error("name not found in user info claims")
	}

	if v, found := claims["picture"]; found && v != nil {
		user.Picture = v.(string)
	} else {
		user.Picture = os.Getenv("SERVER") + "/account-circle.png"
	}

	user.Issuer = idToken.Issuer
	user.IssuedAt = idToken.IssuedAt
	log.Infof("user %s token expires %v", user.Email, idToken.Expiry)

	/* remove auth0 dependency

	domain := os.Getenv("OAUTH2_PROVIDER_URL")
	id := os.Getenv("OAUTH2_CLIENT_ID")
	secret := os.Getenv("OAUTH2_CLIENT_SECRET")
	m, err := management.New(domain, id, secret)
	if err != nil {
		log.Errorf("Error talking to auth0: %v", err)
		// handle err
	}
	u, err := m.User.Read(user.Sub)
	if err != nil {
		log.Errorf("Error reading user %s %v", user.Sub, err)
	} else {

		if u != nil {
			log.Infof("User: %v", u)
			if u.UserMetadata["Plan"] != nil {
				user.Plan = u.UserMetadata["Plan"].(string)
			}
		}

		user.Picture = *u.Picture

		log.Infof("user.Plan: %s", user.Plan)
	}
	*/

	accounts, err := mongodb.ReadAllAccounts(user.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = "Me"
			if user.Name != user.Email {
				account.Name = user.Name
			}
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

		// Set limits regardless of whether they are being enforced
		_, err := core.ReadLimits(accounts[0].Id)
		if err != nil {
			limits_id, err := util.GenerateRandomString(8)
			if err != nil {
				log.Error(err)
			}
			limits_id = "limits-" + limits_id

			limits := &model.Limits{
				Id:          limits_id,
				AccountID:   accounts[0].Id,
				MaxDevices:  core.GetDefaultMaxDevices(),
				MaxNetworks: core.GetDefaultMaxNetworks(),
				MaxMembers:  core.GetDefaultMaxMembers(),
				MaxServices: core.GetDefaultMaxServices(),
				Tolerance:   core.GetDefaultTolerance(),
				UpdatedBy:   user.Email,
				CreatedBy:   user.Email,
				Updated:     time.Now(),
				Created:     time.Now(),
			}
			mongodb.Serialize(limits_id, "id", "limits", limits)
		}
	}
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Picture == "" {
			accounts[i].Picture = user.Picture
			core.UpdateAccount(accounts[i].Id, accounts[i])
		}

	}
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Id == accounts[i].Parent {
			user.AccountID = accounts[i].Id
			user.Picture = accounts[i].Picture
			break
		}
	}
	if user.AccountID == "" {
		user.AccountID = accounts[0].Id
		user.Picture = accounts[0].Picture
	}
	//res, err := collection.InsertOne(ctx, b)

	err = mongodb.UpsertUser(user)
	if err != nil {
		log.Error(err)
	}
	userCache.Set(oauth2Token.AccessToken, user, 0)
	return user, nil
}
