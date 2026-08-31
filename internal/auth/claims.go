package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Email string `json:"email"`
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}