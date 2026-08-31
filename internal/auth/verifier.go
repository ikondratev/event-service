package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("Invalid token")

type Verifier struct {
	key      *rsa.PublicKey
	issuer   string
	audience string
}

func NewVerifier(issuer, audience, publicKeyPath string) (*Verifier, error) {
	raw, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("public key is not PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	return &Verifier{
		key:      rsaKey,
		issuer:   issuer,
		audience: audience,
	}, nil
}


func (v *Verifier) Parse(token string) (Identity, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, fmt.Errorf("unexpected alg")
			}
			return v.key, nil
		},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5*time.Second),
	)

	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !parsed.Valid {
		return Identity{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return Identity{}, ErrInvalidToken
	}

	return Identity{
		Subject: claims.Subject,
		Email:   claims.Email,
		Scopes:  strings.Fields(claims.Scope),
	}, nil
}