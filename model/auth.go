package model

import "golang.org/x/oauth2"

// Auth structure
type Auth struct {
	Oauth2   bool   `json:"oauth2"`
	ClientId string `json:"clientId"`
	Code     string `json:"code"`
	State    string `json:"state"`
	CodeUrl  string `json:"codeUrl"`
}

// Auth interface to implement as auth provider
type Authentication interface {
	Setup() error
	CodeUrl(state string) string
	Exchange(code string) (*oauth2.Token, error)
	UserInfo(oauth2Token *oauth2.Token) (*User, error)
}
