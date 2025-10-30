package w3c

import "fmt"

// ParseError is returned if there is an error parsing a grammar.
type ParseError struct {
	msg    string
	Line   int
	Offset int
	cause  error
}

// NewParseError instantiates a ParseError.
func NewParseError(msg string, line, offset int, cause error) *ParseError {
	return &ParseError{msg: msg, Line: line, Offset: offset, cause: cause}
}

// Error fulfills the error interface.
func (p *ParseError) Error() string {
	return fmt.Sprintf("parse error on line %d at total offset %d: %s", p.Line, p.Offset, p.msg)
}

// Unwrap allows retrieving the original error (if there is one).
func (p *ParseError) Unwrap() error {
	return p.cause
}

// MarshalError is returned if there is an error marshalling a value.
type MarshalError struct {
	msg   string
	cause error
}

// NewMarshalError instantiates a MarshalError.
func NewMarshalError(msg string, cause error) *MarshalError {
	return &MarshalError{msg: msg, cause: cause}
}

// Error fulfills the error interface.
func (m *MarshalError) Error() string {
	return "marshal error: " + m.msg
}

// Unwrap allows retrieving the original error (if there is one).
func (m *MarshalError) Unwrap() error {
	return m.cause
}
