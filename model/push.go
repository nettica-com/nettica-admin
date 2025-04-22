package model

import (
	"fmt"
	"time"
)

type Push struct {
	Version string `json:"version"                   bson:"version"`
	Id      string `json:"id"                        bson:"id"`
	ApiKey  string `json:"apiKey"                    bson:"apiKey"`
	Title   string `json:"title"                     bson:"title"`
	Message string `json:"message"                   bson:"message"`
	Token   string `json:"token"`
}

// Pusher client/server structure

type Pusher struct {
	Id        string     `json:"id"                        bson:"id"`
	ApiKey    string     `json:"apiKey"                    bson:"apiKey"`
	AccountID string     `json:"accountid"                 bson:"accountid"`
	Server    string     `json:"server"                    bson:"server"`
	Host      string     `json:"host"                      bson:"host"`
	Version   string     `json:"version"                   bson:"version"`
	Enabled   *bool      `json:"enabled,omitempty"         bson:"enabled,omitempty"`
	Created   *time.Time `json:"created"                   bson:"created"`
	Updated   *time.Time `json:"updated"                   bson:"updated"`
}

type PusherInterface interface {
	Load() error
	Register() error
	Send(push *Push) error
	IsValid() []error
}

func (a Push) IsValid() error {
	if a.Id == "" || a.Version == "" || a.ApiKey == "" || a.Title == "" || a.Message == "" || a.Token == "" {
		return fmt.Errorf("all fields are required")
	}
	return nil
}

func (a Pusher) Load() error {
	return nil
}

func (a Pusher) Register() error {
	return nil
}

func (a Pusher) Send(push *Push) error {
	if push == nil {
		return fmt.Errorf("push is nil")
	}
	if err := push.IsValid(); err != nil {
		return err
	}
	// send push notification logic here
	return nil
}

// IsValid check if model is valid
func (a Pusher) IsValid() []error {
	errs := make([]error, 0)

	if a.Version == "" {
		errs = append(errs, fmt.Errorf("version is required. 1.0"))
	}

	if a.Id == "" {
		errs = append(errs, fmt.Errorf("id is required"))
	}

	if a.Server == "" {
		errs = append(errs, fmt.Errorf("server field is required"))
	}

	if a.Host == "" {
		errs = append(errs, fmt.Errorf("host field is required"))
	}

	return errs
}
