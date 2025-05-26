package storage

import "errors"

var (
	ErrURLNotFound = errors.New("URL не найден")
	ErrURLExist    = errors.New("URL существует")
)
