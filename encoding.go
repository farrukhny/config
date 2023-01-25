package config

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// Base64Bytes is a type that can be used to decode base64 encoded
type Base64Bytes []byte

// Decode implements the Decoder interface.
func (b *Base64Bytes) Decode(val string) error {
	val = strings.ReplaceAll(val, "+", "-")
	val = strings.ReplaceAll(val, "/", "_")
	val = strings.ReplaceAll(val, "=", "")

	decoded, err := base64.RawURLEncoding.DecodeString(val)
	if err != nil {
		return err
	}

	*b = decoded
	return nil
}

// Bytes returns the underlying byte slice.
func (b *Base64Bytes) Bytes() []byte {
	return []byte(*b)
}

// HexBytes is a type that can be used to decode hex encoded
type HexBytes []byte

// Decode implements the Decoder interface.
func (h *HexBytes) Decode(val string) error {
	decoded, err := hex.DecodeString(val)
	if err != nil {
		return err
	}

	*h = decoded
	return nil
}

// Bytes returns the underlying byte slice.
func (h *HexBytes) Bytes() []byte {
	return []byte(*h)
}
