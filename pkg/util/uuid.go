package util

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

// New generates a UUIDv7 (RFC 4122 variant, version 7) using
// the current Unix millisecond timestamp and crypto/rand for randomness.
func New() (uuid.UUID, error) {
	var u uuid.UUID

	// 1) 48-bit Unix timestamp in milliseconds
	ms := uint64(time.Now().UnixMilli())
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	var r [10]byte
	if _, err := rand.Read(r[:]); err != nil {
		return uuid.Nil, err
	}

	// set version to 7 (0b0111xxxx)
	u[6] = 0x70 | (r[0] & 0x0F)
	// next 8 bits of randomness
	u[7] = r[1]
	// set variant to RFC4122 (0b10xxxxxx)
	u[8] = 0x80 | (r[2] & 0x3F)
	// remaining 56 bits
	copy(u[9:], r[3:])

	return u, nil
}

// NewUUIDv7 generates a UUIDv7 (RFC 4122 variant, version 7) using New Function
func NewUUIDv7() (string, error) {
	u, err := New()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
