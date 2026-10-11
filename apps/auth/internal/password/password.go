// Package password hashes with argon2id, in the same encoding dyosmos uses.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 19 * 1024
	iterations  = 2
	parallelism = 1
	saltLength  = 16
	keyLength   = 32
	variant     = "argon2id"

	maxMemory      = 128 * 1024
	maxIterations  = 10
	maxParallelism = 4
	maxSaltLength  = 64
	maxKeyLength   = 64
	maxConcurrent  = 4
)

var errMalformed = errors.New("password hash is not a valid argon2id encoding")

// Each hash takes 19 MiB, so a burst of logins can't take the whole machine's memory.
var slots = make(chan struct{}, maxConcurrent)

func Hash(plain string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read password salt: %w", err)
	}
	key := deriveKey(plain, salt, memory, iterations, parallelism, keyLength)
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		variant, argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify reads the cost back out of encoded, so raising it later keeps old hashes verifiable.
func Verify(encoded, plain string) bool {
	salt, want, m, t, p, err := decode(encoded)
	if err != nil {
		return false
	}
	got := deriveKey(plain, salt, m, t, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

var dummyHash = sync.OnceValue(func() string {
	encoded, err := Hash("cka-sim-dummy-verification-subject")
	if err != nil {
		return ""
	}
	return encoded
})

// VerifyDummy spends what a real Verify spends and always fails. Calling it for an unknown email
// makes that case cost the same as a wrong password, so timing can't tell which emails exist.
func VerifyDummy(plain string) bool {
	encoded := dummyHash()
	if encoded == "" {
		return false
	}
	Verify(encoded, plain)
	return false
}

func deriveKey(plain string, salt []byte, m, t uint32, p uint8, length uint32) []byte {
	slots <- struct{}{}
	defer func() { <-slots }()
	return argon2.IDKey([]byte(plain), salt, t, m, p, length)
}

func decode(encoded string) (salt, key []byte, m, t uint32, p uint8, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != variant {
		return nil, nil, 0, 0, 0, errMalformed
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return nil, nil, 0, 0, 0, errMalformed
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return nil, nil, 0, 0, 0, errMalformed
	}
	// The hash decides how much CPU and memory Verify burns, so one we didn't write can't ask for more.
	if m < memory || m > maxMemory || t < iterations || t > maxIterations ||
		p < parallelism || p > maxParallelism {
		return nil, nil, 0, 0, 0, errMalformed
	}
	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < saltLength || len(salt) > maxSaltLength {
		return nil, nil, 0, 0, 0, errMalformed
	}
	key, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) < keyLength || len(key) > maxKeyLength {
		return nil, nil, 0, 0, 0, errMalformed
	}
	return salt, key, m, t, p, nil
}
