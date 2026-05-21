package backend

import "errors"

var (
	InitErr        = errors.New("not initialized")
	AliasErr       = errors.New("alias does not exist")
	NoSetErr       = errors.New("set does not exist")
	DSSizeErr      = errors.New("dataset size too small")
	ProvisionedErr = errors.New("set has already been provisioned")
)
