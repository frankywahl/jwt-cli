package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func TestDecodeWithJWKURL(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	jwkKey, err := jwk.FromRaw(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("build JWK from public key: %v", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, "test-kid"); err != nil {
		t.Fatalf("set kid: %v", err)
	}

	set := jwk.NewSet()
	if err := set.AddKey(jwkKey); err != nil {
		t.Fatalf("add key to set: %v", err)
	}

	jwksBytes, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksBytes)
	}))
	defer srv.Close()

	claims := gojwt.MapClaims{
		"sub":   "integration-test",
		"hello": "world",
		"iat":   time.Now().Unix(),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "test-kid"

	tokenString, err := tok.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	cmd := exec.Command("./"+binaryName, "decode", "--token", tokenString, "--jwk-url", srv.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli error: %v\noutput: %s", err, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("unmarshal output: %v\nraw: %s", err, out)
	}

	if result["signature"] != true {
		t.Errorf("expected signature=true, got %v", result["signature"])
	}

	payload, ok := result["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("payload missing from output")
	}
	if payload["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", payload["hello"])
	}
}

func TestDecodeWithJWKURL_WrongKid(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	jwkKey, err := jwk.FromRaw(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("build JWK: %v", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, "real-kid"); err != nil {
		t.Fatalf("set kid: %v", err)
	}

	set := jwk.NewSet()
	if err := set.AddKey(jwkKey); err != nil {
		t.Fatalf("add key: %v", err)
	}

	jwksBytes, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksBytes)
	}))
	defer srv.Close()

	claims := gojwt.MapClaims{"sub": "test"}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "wrong-kid"

	tokenString, err := tok.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	cmd := exec.Command("./"+binaryName, "decode", "--token", tokenString, "--jwk-url", srv.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected cli error: %v\noutput: %s", err, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("unmarshal output: %v\nraw: %s", err, out)
	}

	if result["signature"] != false {
		t.Errorf("expected signature=false for unverifiable token, got %v", result["signature"])
	}
}
