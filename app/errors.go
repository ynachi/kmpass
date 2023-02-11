package app

import (
	"errors"
)

// Application wide custom error types.
var (
	ErrMinControlNodes = errors.New("control nodes number should be odd and minimum 3")
	ErrMinComputeNodes = errors.New("compute nodes number should be minimum 1")
)
