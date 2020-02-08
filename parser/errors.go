package parser

import (
	"errors"
)

var (
	errUnexpectedEOF   = errors.New("unexpected EOF")
	errUnexpectedToken = errors.New("unexpected token")
)
