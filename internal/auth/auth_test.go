package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "secret-password" {
		t.Fatal("password was stored in plain text")
	}
	if !VerifyPassword(hash, "secret-password") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("wrong password verified")
	}
}

func TestJWTIssueAndVerify(t *testing.T) {
	raw, err := IssueJWT("jwt-secret", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWT: %v", err)
	}
	if err := VerifyJWT("jwt-secret", raw); err != nil {
		t.Fatalf("VerifyJWT: %v", err)
	}
	if err := VerifyJWT("other-secret", raw); err == nil {
		t.Fatal("expected wrong secret to fail")
	}
}
