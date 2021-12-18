package util

import (
	"errors"

	"github.com/satori/go.uuid"
)

var (
	ErrInvalidUUIDString = errors.New("invalid uuid string format")
	InvalidUUID          = UUID{0}
)

type UUID uuid.UUID

func UUIDGen() (UUID, error) {
	id := uuid.NewV4()
	return UUID(id), nil
}

func (id UUID) String() string {
	return uuid.UUID(id).String()
}

func UUIDFromString(s string) (UUID, error) {
	id, err := uuid.FromString(s)
	if err != nil {
		return InvalidUUID, err
	}

	return UUID(id), nil
}

func UUIDStringGen() (string, error) {
	id := uuid.NewV4()
	return id.String(), nil
}
