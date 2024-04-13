package internal

import "errors"

var (
	ErrFileOpen error = errors.New("error opening file")
	ErrFileRead error = errors.New("error reading file")
)
