package auth

import (
	"testing"
	"time"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	secret := []byte("test-super-secret-key-32-bytes!!")

	claims := Claims{
		Subject:      "018f0000-0000-7000-8000-000000000001",
		Email:        "investor@example.com",
		DisplayName:  "Test Investor",
		BaseCurrency: "USD",
		IssuedAt:     time.Now().Unix(),
		ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
	}

	token, err := GenerateAccessToken(secret, claims)
	if err != nil {
		t.Fatalf("unexpected error generating access token: %v", err)
	}

	parsed, err := ValidateAccessToken(secret, token)
	if err != nil {
		t.Fatalf("unexpected error validating valid token: %v", err)
	}

	if parsed.Subject != claims.Subject {
		t.Errorf("expected subject %s, got %s", claims.Subject, parsed.Subject)
	}
	if parsed.Email != claims.Email {
		t.Errorf("expected email %s, got %s", claims.Email, parsed.Email)
	}

	// Test invalid secret signature rejection
	wrongSecret := []byte("wrong-secret-key-that-should-fail")
	_, err = ValidateAccessToken(wrongSecret, token)
	if err == nil {
		t.Errorf("expected validation to fail with wrong secret, got nil")
	}

	// Test expired token rejection
	expiredClaims := claims
	expiredClaims.ExpiresAt = time.Now().Add(-1 * time.Minute).Unix()
	expiredToken, err := GenerateAccessToken(secret, expiredClaims)
	if err != nil {
		t.Fatalf("unexpected error generating expired token: %v", err)
	}

	_, err = ValidateAccessToken(secret, expiredToken)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	rawToken, hashHex, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}

	if len(rawToken) != 64 { // 32 bytes hex = 64 hex chars
		t.Errorf("expected raw token length 64, got %d", len(rawToken))
	}

	expectedHash := HashRefreshToken(rawToken)
	if hashHex != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, hashHex)
	}
}
