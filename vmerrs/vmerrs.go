package vmerrs

import (
	"errors"
)

// List evm execution errors
var (
	ErrWriteProtection = errors.New("write protection")
)
