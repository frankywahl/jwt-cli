package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

func TestDecodeExpiredToken(t *testing.T) {
	claims := gojwt.MapClaims{
		"sub": "test",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	tokenString, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	cmd := exec.Command("./"+binaryName, "decode", "--secret", "secret", "--token", tokenString)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for expired token, got exit 0")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal JSON stdout: %v\nraw: %s", err, stdout.String())
	}
	if result["active"] != false {
		t.Errorf("expected active=false, got %v", result["active"])
	}

	if !strings.Contains(stderr.String(), "WARNING: token is expired") {
		t.Errorf("expected expiry warning on stderr, got: %q", stderr.String())
	}
}

func TestDecodeNonExpiredToken(t *testing.T) {
	claims := gojwt.MapClaims{
		"sub": "test",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	tokenString, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	cmd := exec.Command("./"+binaryName, "decode", "--secret", "secret", "--token", tokenString)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error for valid token: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "WARNING") {
		t.Errorf("unexpected warning for non-expired token: %s", stderr.String())
	}
}

func TestDecodeNoExpClaim(t *testing.T) {
	claims := gojwt.MapClaims{
		"sub": "test",
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	tokenString, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	cmd := exec.Command("./"+binaryName, "decode", "--secret", "secret", "--token", tokenString)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error for token without exp: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "WARNING") {
		t.Errorf("unexpected warning for token with no exp claim: %s", stderr.String())
	}
}
