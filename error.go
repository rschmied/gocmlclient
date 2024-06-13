package cmlclient

import "errors"

var (
	ErrSystemNotReady       = errors.New("system not ready")
	ErrElementNotFound      = errors.New("element not found")
	ErrNoNamedConfigSupport = errors.New("backend does not support named configs")
)
