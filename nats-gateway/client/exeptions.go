package client

import (
	"errors"
)

var (
	ErrMaxMessageSize    = errors.New("message size exceeds max message size")
	ErrArgsNotProtoType  = errors.New("args not proto type")
	ErrReplyNotProtoType = errors.New("reply not proto type")
)
