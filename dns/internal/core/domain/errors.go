package domain

import "errors"

var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrAlreadyExists  = errors.New("entity already exists")
)
