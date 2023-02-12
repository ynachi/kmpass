package app

import (
	"errors"
)

// Application wide custom error types.
var (
	ErrMinControlNodes  = errors.New("control nodes number should be odd and minimum 3")
	ErrMinComputeNodes  = errors.New("compute nodes number should be minimum 1")
	ErrGetHomeDirectory = errors.New("cannot get user's home directory")
	ErrLoadTemplate     = errors.New("cannot load template from embed")
	ErrCreateFile       = errors.New("cannot create new file")
	ErrParseTemplate    = errors.New("cannot parse template")
	ErrBase64Encode     = errors.New("cannot encode file to b64")
)
