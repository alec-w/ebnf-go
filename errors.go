package ebnf

import "fmt"

type JsonError struct {
	wrapped error
}

func (j *JsonError) Error() string {
	return j.wrapped.Error()
}

func (j *JsonError) Unwrap() error {
	return j.wrapped
}

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
