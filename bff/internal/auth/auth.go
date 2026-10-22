package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	UserRoleKey       contextKey = "user_role"
	ClaimsKey         contextKey = "claims"
	ResponseWriterKey contextKey = "http_response_writer"
	RequestKey        contextKey = "http_request"
)

const (
	RefreshCookieName   = "graphfolio_refresh"
	DefaultDevJWTSecret = "graphfolio-default-dev-secret-key-256bit-ok!"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrInvalidToken     = errors.New("invalid or expired token")
)

type Claims struct {
	Subject      string `json:"sub"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	BaseCurrency string `json:"base_currency"`
	Role         string `json:"role"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
}

// GetJWTSecret returns the secret key configured in JWT_SECRET or the default dev key.
func GetJWTSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte(DefaultDevJWTSecret)
}

// ValidateAccessToken verifies the token signature and expiration, returning parsed Claims.
func ValidateAccessToken(secret []byte, tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsignedPayload := parts[0] + "." + parts[1]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsignedPayload))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return nil, ErrInvalidToken
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// GetUserID retrieves the authenticated user ID from context.
func GetUserID(ctx context.Context) (string, error) {
	if val, ok := ctx.Value(UserIDKey).(string); ok && val != "" {
		return val, nil
	}
	return "", ErrUnauthenticated
}

// GetUserRole retrieves the authenticated user's role from context.
func GetUserRole(ctx context.Context) (string, error) {
	if val, ok := ctx.Value(UserRoleKey).(string); ok && val != "" {
		return val, nil
	}
	return "", ErrUnauthenticated
}

// SetRefreshCookie writes the httpOnly SameSite=Strict refresh token cookie.
func SetRefreshCookie(ctx context.Context, token string, maxAgeSeconds int) {
	if w, ok := ctx.Value(ResponseWriterKey).(http.ResponseWriter); ok {
		http.SetCookie(w, &http.Cookie{
			Name:     RefreshCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   maxAgeSeconds,
		})
	}
}

// ClearRefreshCookie removes the refresh token cookie.
func ClearRefreshCookie(ctx context.Context) {
	if w, ok := ctx.Value(ResponseWriterKey).(http.ResponseWriter); ok {
		http.SetCookie(w, &http.Cookie{
			Name:     RefreshCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
	}
}

// GetRefreshCookie reads the refresh token from the HTTP request cookie.
func GetRefreshCookie(ctx context.Context) (string, error) {
	if r, ok := ctx.Value(RequestKey).(*http.Request); ok {
		cookie, err := r.Cookie(RefreshCookieName)
		if err != nil {
			return "", err
		}
		return cookie.Value, nil
	}
	return "", fmt.Errorf("http request not available in context")
}

// Middleware handles token parsing and context injection for HTTP requests.
func Middleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ResponseWriterKey, w)
			ctx = context.WithValue(ctx, RequestKey, r)

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := ValidateAccessToken(secret, tokenStr)
				if err == nil && claims.Subject != "" {
					ctx = context.WithValue(ctx, UserIDKey, claims.Subject)
					if claims.Role != "" {
						ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
					}
					ctx = context.WithValue(ctx, ClaimsKey, claims)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

