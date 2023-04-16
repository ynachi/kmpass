package app

import (
	"errors"
)

// Application wide custom error types.
var (
	ErrMinControlNodes      = errors.New("control nodes number should be odd and minimum 3")
	ErrMinComputeNodes      = errors.New("compute nodes number should be minimum 1")
	ErrGetHomeDirectory     = errors.New("cannot get user's home directory")
	ErrLoadTemplate         = errors.New("cannot load template from embed")
	ErrCreateFile           = errors.New("cannot create new file")
	ErrParseTemplate        = errors.New("cannot parse template")
	ErrBase64Encode         = errors.New("cannot encode file to b64")
	ErrMemFormat            = errors.New("memory format should be like <number>K|M|G|k|m|g")
	ErrInvalidIPV4Address   = errors.New("invalid IP address provided")
	ErrVMNotRunning         = errors.New("VM is not running")
	ErrVMAlreadyExist       = errors.New("VM already exist")
	ErrCloudInitGeneration  = errors.New("unable to generate cloud init file")
	ErrClusterConfiguration = errors.New("wrong value in cluster configuration")
	ErrOddNumberCtrlNode    = errors.New("number of control nodes should be odd")
)
