package solana

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

// base58Alphabet is the Bitcoin/Solana base58 alphabet.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Decode decodes a base58-encoded string into raw bytes.
func base58Decode(input string) ([]byte, error) {
	if input == "" {
		return nil, errors.New("empty input")
	}

	// Count leading '1's (zero bytes in base58).
	leadingZeros := 0
	for i := 0; i < len(input) && input[i] == '1'; i++ {
		leadingZeros++
	}

	// Convert base58 string to a big integer.
	result := big.NewInt(0)
	for i := 0; i < len(input); i++ {
		digit := bytes.IndexByte([]byte(base58Alphabet), input[i])
		if digit < 0 {
			return nil, fmt.Errorf("invalid base58 character %q", input[i])
		}
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(digit)))
	}

	// Convert the big integer to bytes.
	decoded := result.Bytes()

	// Prepend zero bytes for leading '1's.
	if leadingZeros > 0 {
		prefix := make([]byte, leadingZeros)
		decoded = append(prefix, decoded...)
	}

	return decoded, nil
}

// base58Encode encodes raw bytes into a base58-encoded string.
func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	// Count leading zero bytes.
	leadingZeros := 0
	for i := 0; i < len(input) && input[i] == 0; i++ {
		leadingZeros++
	}

	// Convert the big-endian bytes to a big integer.
	result := new(big.Int).SetBytes(input)

	// Repeatedly divide by 58 to produce base58 digits.
	var digits []byte
	for result.Sign() > 0 {
		remainder := new(big.Int)
		result.DivMod(result, big.NewInt(58), remainder)
		digits = append([]byte{base58Alphabet[remainder.Int64()]}, digits...)
	}

	// Prepend '1' for each leading zero byte.
	for i := 0; i < leadingZeros; i++ {
		digits = append([]byte{'1'}, digits...)
	}

	return string(digits)
}

// MarshalBinary serializes a V0 Message to wire format.
func (m *Message) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Header: 3 bytes.
	buf.WriteByte(m.Header.NumRequiredSignatures)
	buf.WriteByte(m.Header.NumReadonlySignedAccounts)
	buf.WriteByte(m.Header.NumReadonlyUnsignedAccounts)

	// Account keys count (1-byte uint8, compact uint).
	if len(m.AccountKeys) > 255 {
		return nil, errors.New("too many account keys for V0 message")
	}
	buf.WriteByte(byte(len(m.AccountKeys)))
	for _, key := range m.AccountKeys {
		buf.Write(key[:])
	}

	// Recent blockhash (32 bytes).
	buf.Write(m.RecentBlockhash[:])

	// Instructions count (1-byte uint8).
	if len(m.Instructions) > 255 {
		return nil, errors.New("too many instructions for V0 message")
	}
	buf.WriteByte(byte(len(m.Instructions)))
	for _, instr := range m.Instructions {
		// Program ID index.
		buf.WriteByte(instr.ProgramIDIndex)
		// Account indexes count.
		if len(instr.Accounts) > 255 {
			return nil, errors.New("too many account indexes")
		}
		buf.WriteByte(byte(len(instr.Accounts)))
		for _, acc := range instr.Accounts {
			buf.WriteByte(acc)
		}
		// Data length (compact uint: 1 byte for <253).
		if len(instr.Data) > 255 {
			return nil, errors.New("instruction data too long for simple encoding")
		}
		buf.WriteByte(byte(len(instr.Data)))
		buf.Write(instr.Data)
	}

	return buf.Bytes(), nil
}

// EncodeUint64LE appends a little-endian uint64 to buf.
func EncodeUint64LE(buf *bytes.Buffer, v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	buf.Write(b[:])
}

// DecodeUint64LE reads a little-endian uint64 from b.
func DecodeUint64LE(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
