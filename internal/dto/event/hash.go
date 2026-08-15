package eventdto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Hash(req CreateRequest) (string, error) {
	raw, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}