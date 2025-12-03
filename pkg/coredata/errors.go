package coredata

import (
	"errors"
)

var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrResourceInUse         = errors.New("resource is in use")
)
