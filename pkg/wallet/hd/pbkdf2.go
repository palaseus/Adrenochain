package hd

import (
	"crypto/hmac"
	"crypto/sha512"
)

// pbkdf2HMACSHA512 derives a key using PBKDF2 with HMAC-SHA512.
func pbkdf2HMACSHA512(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	if iterations < 1 {
		return nil, errInvalidIterations
	}
	if keyLen < 1 {
		return nil, errInvalidKeyLen
	}

	prf := func(key, data []byte) []byte {
		h := hmac.New(sha512.New, key)
		h.Write(data)
		return h.Sum(nil)
	}

	hashLen := sha512.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var result []byte
	for i := 1; i <= numBlocks; i++ {
		block := prf(password, append(salt, uint32Bytes(uint32(i))...))
		u := make([]byte, len(block))
		copy(u, block)
		for j := 2; j <= iterations; j++ {
			block = prf(password, block)
			for k := range u {
				u[k] ^= block[k]
			}
		}
		result = append(result, u...)
	}
	return result[:keyLen], nil
}

func uint32Bytes(v uint32) []byte {
	return []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}

// pad32 left-pads a big-endian byte slice to 32 bytes.
func pad32(b []byte) []byte {
	if len(b) >= 32 {
		return b
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

var (
	errInvalidIterations = &hdError{"iterations must be >= 1"}
	errInvalidKeyLen     = &hdError{"key length must be >= 1"}
)

type hdError struct {
	msg string
}

func (e *hdError) Error() string { return e.msg }
