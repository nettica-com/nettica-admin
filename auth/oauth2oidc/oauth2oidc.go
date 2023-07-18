package oauth2oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"github.com/coreos/go-oidc"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	mongodb "github.com/nettica-com/nettica-admin/mongo"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"gopkg.in/auth0.v4/management"

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
)

// Setup validate provider
func (o *Oauth2idc) Setup() error {
	var err error

	userCache = cache.New(60*time.Minute, 10*time.Minute)

	oidcProvider, err = oidc.NewProvider(context.TODO(), os.Getenv("OAUTH2_PROVIDER"))
	if err != nil {
		return err
	}

	oidcIDTokenVerifier = make([]*oidc.IDTokenVerifier, 0)
	oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID")}))
	oidcIDTokenVerifier = append(oidcIDTokenVerifier, oidcProvider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID_WINDOWS")}))

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
func (o *Oauth2idc) CodeUrl(state string) string {
	return oauth2Config.AuthCodeURL(state)
}

// Exchange exchange code for Oauth2 token
func (o *Oauth2idc) Exchange(code string) (*oauth2.Token, error) {
	oauth2Token, err := oauth2Config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

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
		user.Picture = "/account-circle.svg"
	}

	user.Issuer = idToken.Issuer
	user.IssuedAt = idToken.IssuedAt
	log.Infof("user %s token expires %v", user.Email, idToken.Expiry)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection("users")

	data, err := json.Marshal(user)
	//	json := fmt.Sprintf("%v", user)
	var b interface{}
	err = bson.UnmarshalExtJSON([]byte(data), true, &b)

	findstr := fmt.Sprintf("{\"email\":\"%s\"}", user.Email)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	update := bson.M{
		"$set": b,
	}

	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error(err)
	}

	accounts, err := mongodb.ReadAllAccounts(user.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Email = user.Email
			account.Role = "Owner"
			account.Status = "Active"
			a, err := core.CreateAccount(&account)
			log.Infof("CREATE ACCOUNT = %v", a)
			if err != nil {
				log.Error(err)
			}
		}
	}

	//res, err := collection.InsertOne(ctx, b)

	log.Infof("Res: %v", res)
	userCache.Set(oauth2Token.AccessToken, user, 0)
	return user, nil
}
