package feralfilefeature

import "errors"

var (
	ErrInvalidAddress   = errors.New("Invalid address provided")
	ErrInvalidPublicKey = errors.New("Invalid public key provided")
	ErrInvalidSignature = errors.New("Invalid signature provided")
	ErrInvalidTokenID   = errors.New("Invalid tokenID provided")
)
