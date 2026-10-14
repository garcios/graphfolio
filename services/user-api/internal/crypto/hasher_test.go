package crypto

import (
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	// Use slightly smaller memory parameters for fast test execution
	testParams := Argon2Params{
		Memory:      8 * 1024, // 8MB
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	password := "SecretPass#2026!"

	hash, err := HashPassword(password, testParams)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$m=8192,t=1,p=2$") {
		t.Errorf("unexpected hash format: %s", hash)
	}

	// Verify correct password
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("unexpected error verifying password: %v", err)
	}
	if !match {
		t.Errorf("expected password to match, got false")
	}

	// Verify incorrect password
	match, err = VerifyPassword("WrongPassword123!", hash)
	if err != nil {
		t.Fatalf("unexpected error verifying incorrect password: %v", err)
	}
	if match {
		t.Errorf("expected password not to match, got true")
	}

	// Verify malformed hash returns error
	_, err = VerifyPassword(password, "malformed-hash")
	if err == nil {
		t.Errorf("expected error for malformed hash, got nil")
	}
}

func TestDevSeedPassword(t *testing.T) {
	seedHash := "$argon2id$v=19$m=65536,t=3,p=4$z48VmX+qFO9X2FkAfV/Z6Q$WW1NOG+O6TwEwlSDkDIfae3Xi0xOePDpNg+rd7lkWsw"
	match, err := VerifyPassword("Password123!", seedHash)
	if err != nil {
		t.Fatalf("unexpected error verifying seed password: %v", err)
	}
	if !match {
		t.Errorf("expected Password123! to match seed hash")
	}

	wrongMatch, err := VerifyPassword("password123", seedHash)
	if err != nil {
		t.Fatalf("unexpected error verifying wrong password: %v", err)
	}
	if wrongMatch {
		t.Errorf("expected lowercase password123 not to match seed hash")
	}
}

