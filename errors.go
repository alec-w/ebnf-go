package ebnf

import "fmt"

// JSONError is returned if there is an error marshalling a Syntax as JSON.
type JSONError struct {
	wrapped error
}

func (j *JSONError) Error() string {
	return j.wrapped.Error()
}

func (j *JSONError) Unwrap() error {
	return j.wrapped
}

// ParseError is returned if there is an error parsing an EBNF grammar.
type ParseError struct {
	Msg     string
	wrapped error
	Offset  int
}

func (p *ParseError) Error() string {
	return fmt.Sprintf("Parse error at offset %d: %s", p.Offset, p.Msg)
}

func (p *ParseError) Unwrap() error {
	return p.wrapped
}
