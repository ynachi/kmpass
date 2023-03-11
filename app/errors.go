package app

import (
	"errors"
)

// Application wide custom error types.
var (
	ErrMinControlNodes     = errors.New("control nodes number should be odd and minimum 3")
	ErrMinComputeNodes     = errors.New("compute nodes number should be minimum 1")
	ErrGetHomeDirectory    = errors.New("cannot get user's home directory")
	ErrLoadTemplate        = errors.New("cannot load template from embed")
	ErrCreateFile          = errors.New("cannot create new file")
	ErrParseTemplate       = errors.New("cannot parse template")
	ErrBase64Encode        = errors.New("cannot encode file to b64")
	ErrMinDiskSize         = errors.New("disk size cannot be less than required minimum")
	ErrMinMemSize          = errors.New("memory size cannot be less than required minimum")
	ErrMemFormat           = errors.New("memory format should be like <number>K|M|G|k|m|g")
	ErrInvalidIPV4Address  = errors.New("invalid IP address provided")
	ErrInvalidCoreFmt      = errors.New("invalid number of core, integer required")
	ErrVMNotExist          = errors.New("vm does not exist")
	ErrVMNotRunning        = errors.New("VM is not running")
	ErrVMAlreadyExist      = errors.New("VM already exist")
	ErrCloudInitGeneration = errors.New("unable to generate cloud init file")
)
