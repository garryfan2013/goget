package util

import (
	"errors"

	"github.com/satori/go.uuid"
)

var (
	ErrInvalidUUIDString = errors.New("Invalid uuid string format")
	InvalidUUID          = UUID{0}
)

type UUID uuid.UUID

func UUIDGen() (UUID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return InvalidUUID, err
	}

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
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
