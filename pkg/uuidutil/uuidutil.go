package uuidutil

import (
	"fmt"

	"github.com/google/uuid"
)

func Parse(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid uuid %q: %w", s, err)
	}
	return id, nil
}
