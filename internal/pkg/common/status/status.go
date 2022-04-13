package status

import (
	"fmt"
	"howdah/internal/pkg/common/codes"
)

// Mocking Errors

type Status struct {
	code    int32
	message string
}

func New(c codes.Code, msg string) *Status {
	return &Status{
		code:    int32(c),
		message: msg,
	}
}

func Errorf(c codes.Code, format string, a ...interface{}) error {
	return Err(c, fmt.Sprintf(format, a...))
}

func (s *Status) Err() error {
	return &Error{s: s}
}

func (s *Status) Code() codes.Code {
	return codes.Code(s.code)
}

func (s *Status) Message() string {
	return s.message
}

func (s *Status) String() string {
	return fmt.Sprintf("howdah error: code = %s desc = %s", s.Code(), s.Message())
}

func Err(c codes.Code, msg string) error {
	return New(c, msg).Err()
}

type Error struct {
	s *Status
}

func (e *Error) Error() string {
	return e.s.String()
}