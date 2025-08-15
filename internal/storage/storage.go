package storage

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrDB                = errors.New("database error")
	ErrNoUsers           = errors.New("no users found")
)
