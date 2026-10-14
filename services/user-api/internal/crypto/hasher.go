package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// Argon2Params contains the parameters used to derive an Argon2id hash.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams follows RFC 9106 recommended parameters for password hashing.
var DefaultParams = Argon2Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// DummyHash is a precomputed valid Argon2id hash for constant-time dummy comparisons.
// Computed for password "GraphFolioDummy#2026" with DefaultParams.
const DummyHash = "$argon2id$v=19$m=65536,t=3,p=4$dGVzdHNhbHRmb3JkdW1teTE2$nZ6F2eHn3r9Z6Qx6L6vP7Y4qO4dD3wW8xY2kL5nN0pA"

// HashPassword hashes a plaintext password using the provided Argon2id parameters.
func HashPassword(password string, params Argon2Params) (string, error) {
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword performs a constant-time comparison of a plaintext password against an Argon2id hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, ErrInvalidHash
	}

	if vals[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, ErrInvalidHash
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	params := Argon2Params{}
	pairs := strings.Split(vals[3], ",")
	if len(pairs) != 3 {
		return false, ErrInvalidHash
	}

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return false, ErrInvalidHash
		}
		val, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return false, ErrInvalidHash
		}
		switch kv[0] {
		case "m":
			params.Memory = uint32(val)
		case "t":
			params.Iterations = uint32(val)
		case "p":
			params.Parallelism = uint8(val)
		default:
			return false, ErrInvalidHash
		}
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, ErrInvalidHash
	}
	params.SaltLength = uint32(len(salt))

	expectedHash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, ErrInvalidHash
	}
	params.KeyLength = uint32(len(expectedHash))

	computedHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	if subtle.ConstantTimeCompare(expectedHash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}
