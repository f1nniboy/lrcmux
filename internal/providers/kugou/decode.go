package kugou

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var krcKey = [16]byte{64, 71, 97, 119, 94, 50, 116, 71, 81, 54, 49, 45, 206, 210, 110, 105}

func decodeKRC(content string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", fmt.Errorf("krc base64: %w", err)
	}
	if len(raw) < 4 {
		return "", errors.New("krc too short")
	}
	encrypted := raw[4:]
	decrypted := make([]byte, len(encrypted))
	for i, b := range encrypted {
		decrypted[i] = b ^ krcKey[i%len(krcKey)]
	}
	r, err := zlib.NewReader(bytes.NewReader(decrypted))
	if err != nil {
		return "", fmt.Errorf("krc zlib: %w", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("krc decompress: %w", err)
	}
	return string(out), nil
}
