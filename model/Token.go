package model

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// Token model
type Token struct {
	UserID string `json:"userId"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.StandardClaims
}
