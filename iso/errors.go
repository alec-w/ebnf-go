package iso

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

// ParseRuleError is returned to indicate which rule was being parsed when an error occurred parsing an EBNF grammar.
// Receivers of this error should unwrap it to get more detail on the original parse error itself.
type ParseRuleError struct {
	MetaIdentifier string
	Line           int
	Wrapped        *ParseError
}

func (p *ParseRuleError) Error() string {
	if p.MetaIdentifier != "" {
		return fmt.Sprintf("failed parsing rule %s beginning on line %d", p.MetaIdentifier, p.Line)
	}

	return fmt.Sprintf("failed parsing rule beginning on line %d", p.Line)
}

func (p *ParseRuleError) Unwrap() error {
	return p.Wrapped
}

// ParseError is returned if there is an error parsing an EBNF grammar.
type ParseError struct {
	Msg     string
	wrapped error
	Offset  int
	Line    int
}

func (p *ParseError) Error() string {
	return fmt.Sprintf("parse error at offset %d: %s", p.Offset, p.Msg)
}

func (p *ParseError) Unwrap() error {
	return p.wrapped
}
