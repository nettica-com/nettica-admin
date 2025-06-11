package model

import "time"

// RefreshToken represents a stored OAuth2 refresh token and its metadata
// Place this file as model/refresh_token.go

type RefreshToken struct {
	Token     string    `json:"token"      bson:"token"`
	Sub       string    `json:"sub"        bson:"sub"`
	Email     string    `json:"email"      bson:"email"`
	IssuedAt  time.Time `json:"issued_at"  bson:"issued_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	Revoked   bool      `json:"revoked"    bson:"revoked"`
}
