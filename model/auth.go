package model

import "golang.org/x/oauth2"

// Auth structure
type Auth struct {
	Oauth2     bool     `json:"oauth2"`
	ClientId   string   `json:"clientId"`
	Code       string   `json:"code"`
	State      string   `json:"state"`
	CodeUrl    string   `json:"codeUrl"`
	Redirect   string   `json:"redirect_uri"`
	Audience   string   `json:"audience"`
	Connection string   `json:"connection"`
	IdToken    string   `json:"id_token"`
	Providers  []string `json:"providers"`
}

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Expiry       string `json:"expiry,omitempty"`
}

// Auth interface to implement as auth provider
type Authentication interface {
	Setup() error
	CodeUrl(state string) string
	CodeUrl2(state string) string
	Exchange(auth Auth) (*oauth2.Token, error)
	Exchange2(code string) (*oauth2.Token, error)
	UserInfo(oauth2Token *oauth2.Token) (*User, error)
}
