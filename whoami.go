package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// whoami verifies the platform's identity header with the public key the
// platform injected (DEPLOYMIND_IDENTITY_KEY) and echoes the claims. This is
// what the template middleware will do for every app in Beta.
func whoami(w http.ResponseWriter, r *http.Request) {
	raw := r.Header.Get("X-DeployMind-Identity")
	if raw == "" {
		http.Error(w, "no identity header", 401)
		return
	}
	pub, err := parseKey(os.Getenv("DEPLOYMIND_IDENTITY_KEY"))
	if err != nil {
		http.Error(w, "bad key: "+err.Error(), 500)
		return
	}
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pub, nil
	}, jwt.WithAudience(os.Getenv("DEPLOYMIND_APP")), jwt.WithIssuer("deploymind"))
	if err != nil {
		http.Error(w, "invalid identity: "+err.Error(), 401)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"verified": true, "email": claims["email"], "role": claims["role"], "org": claims["org"], "app": claims["app"]})
}

func parseKey(s string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, errors.New("not PEM")
	}
	k, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := k.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not ed25519")
	}
	return pub, nil
}
