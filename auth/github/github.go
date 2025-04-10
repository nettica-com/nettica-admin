package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	"golang.org/x/oauth2"
	oauth2Github "golang.org/x/oauth2/github"
)

// Github in order to implement interface, struct is required
type Github struct{}

var (
	oauth2Config *oauth2.Config
)

// Setup validate provider
func (o *Github) Setup() error {
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:       []string{"user"},
		Endpoint:     oauth2Github.Endpoint,
	}

	return nil
}

// CodeUrl get url to redirect client for auth
func (o *Github) CodeUrl(state string) string {
	return oauth2Config.AuthCodeURL(state)
}

func (o *Github) CodeUrl2(state string) string {
	return o.CodeUrl(state)
}

// Exchange exchange code for Oauth2 token
func (o *Github) Exchange(auth model.Auth) (*oauth2.Token, error) {
	oauth2Token, err := oauth2Config.Exchange(context.TODO(), auth.Code)
	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}

func (o *Github) Exchange2(code string) (*oauth2.Token, error) {
	token, err := o.Exchange(model.Auth{Code: code})
	return token, err
}

// UserInfo get token user
func (o *Github) UserInfo(oauth2Token *oauth2.Token) (*model.User, error) {
	// https://developer.github.com/apps/building-oauth-apps/authorizing-oauth-apps/

	// we have the token, lets get user information
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", oauth2Token.AccessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status %s expect 200 OK", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}

	// get some infos about user
	user := &model.User{}

	if val, ok := data["name"]; ok && val != nil {
		user.Name = val.(string)
	}
	if val, ok := data["email"]; ok && val != nil {
		user.Email = val.(string)
	}
	if val, ok := data["html_url"]; ok && val != nil {
		user.Profile = val.(string)
	}

	// openid specific
	user.Sub = "github is not an openid provider"
	user.Issuer = "https://github.com"
	user.IssuedAt = time.Now().UTC()

	return user, nil
}
