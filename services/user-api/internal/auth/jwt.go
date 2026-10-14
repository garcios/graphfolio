package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token format or signature")
	ErrTokenExpired = errors.New("token has expired")
)

// Header represents the standard JOSE header for HS256 JWTs.
type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Claims represents standard and custom GraphFolio claims embedded in access tokens.
type Claims struct {
	Subject      string `json:"sub"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	BaseCurrency string `json:"base_currency"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
}

// GenerateAccessToken signs a JWT using HMAC-SHA256.
func GenerateAccessToken(secret []byte, claims Claims) (string, error) {
	header := Header{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwt claims: %w", err)
	}

	headerEnc := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEnc := base64.RawURLEncoding.EncodeToString(claimsJSON)

	unsignedPayload := headerEnc + "." + claimsEnc

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsignedPayload))
	signature := mac.Sum(nil)
	sigEnc := base64.RawURLEncoding.EncodeToString(signature)

	return unsignedPayload + "." + sigEnc, nil
}

// ValidateAccessToken verifies the token signature and expiration, returning parsed Claims.
func ValidateAccessToken(secret []byte, tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerEnc := parts[0]
	claimsEnc := parts[1]
	sigEnc := parts[2]

	unsignedPayload := headerEnc + "." + claimsEnc

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsignedPayload))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(sigEnc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return nil, ErrInvalidToken
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(claimsEnc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// GenerateRefreshToken generates a secure 32-byte random token and its SHA-256 hash.
func GenerateRefreshToken() (rawToken string, hashHex string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	rawToken = hex.EncodeToString(buf)
	hashHex = HashRefreshToken(rawToken)

	return rawToken, hashHex, nil
}

// HashRefreshToken computes the SHA-256 hex digest of a raw refresh token string.
func HashRefreshToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}
