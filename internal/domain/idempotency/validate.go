package idempotency

import (
	"strings"

	"github.com/google/uuid"
)

func ValidateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrMissingKey
	} else if err := uuid.Validate(key); err != nil {
		return ErrInvalidKey
	}
	
	return nil
}