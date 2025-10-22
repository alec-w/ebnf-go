package w3c

import "encoding/json"

type Syntax struct {
	Rules []Rule
}

type Rule struct {
	Line       int
	Symbol     string
	Expression Expression
}

type Expression interface {
	ListExpression() *ListExpression
	AlternateExpression() *AlternateExpression
	ExceptionExpression() *ExceptionExpression
	SymbolExpression() *SymbolExpression
	CharacterSetExpression() *CharacterSetExpression
	LiteralExpression() *LiteralExpression
	Optional() bool
	OneOrMore() bool
	ZeroOrMore() bool
	setOptional(bool)
	setOneOrMore(bool)
	setZeroOrMore(bool)
	isParenthesised() bool
	setParenthesised(bool)
}

var _ Expression = &baseExpression{}

type baseExpression struct {
	optional      bool
	oneOrMore     bool
	zeroOrMore    bool
	parenthesised bool
}

func (b *baseExpression) ListExpression() *ListExpression {
	return nil
}

func (b *baseExpression) AlternateExpression() *AlternateExpression {
	return nil
}

func (b *baseExpression) ExceptionExpression() *ExceptionExpression {
	return nil
}

func (b *baseExpression) SymbolExpression() *SymbolExpression {
	return nil
}

func (b *baseExpression) CharacterSetExpression() *CharacterSetExpression {
	return nil
}

func (b *baseExpression) LiteralExpression() *LiteralExpression {
	return nil
}

func (b *baseExpression) Optional() bool {
	return b.optional
}

func (b *baseExpression) OneOrMore() bool {
	return b.oneOrMore
}

func (b *baseExpression) ZeroOrMore() bool {
	return b.zeroOrMore
}

func (b *baseExpression) setOptional(optional bool) {
	b.optional = optional
}

func (b *baseExpression) setOneOrMore(oneOrMore bool) {
	b.optional = oneOrMore
}

func (b *baseExpression) setZeroOrMore(zeroOrMore bool) {
	b.optional = zeroOrMore
}

func (b *baseExpression) isParenthesised() bool {
	return b.parenthesised
}

func (b *baseExpression) setParenthesised(parenthesised bool) {
	b.parenthesised = parenthesised
}

var _ Expression = &ListExpression{}

type ListExpression struct {
	baseExpression
	Expressions []Expression
}

func (l *ListExpression) ListExpression() *ListExpression {
	return l
}

func (l *ListExpression) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":        "list",
		"expressions": l.Expressions,
	})
}

var _ Expression = &AlternateExpression{}

type AlternateExpression struct {
	baseExpression
	Expressions []Expression
}

func (a *AlternateExpression) AlternateExpression() *AlternateExpression {
	return a
}

func (a *AlternateExpression) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":        "alternate",
		"expressions": a.Expressions,
	})
}

var _ Expression = &ExceptionExpression{}

type ExceptionExpression struct {
	baseExpression
	Match  Expression
	Except Expression
}

func (e *ExceptionExpression) ExceptionExpression() *ExceptionExpression {
	return e
}

var _ Expression = &SymbolExpression{}

type SymbolExpression struct {
	baseExpression
	Symbol string
}

func (e *SymbolExpression) SymbolExpression() *SymbolExpression {
	return e
}

var _ Expression = &CharacterSetExpression{}

type CharacterSetExpression struct {
	baseExpression
	Enumerations []string
	Ranges       []Range
	Forbidden    bool
}

func (c *CharacterSetExpression) CharacterSetExpression() *CharacterSetExpression {
	return c
}

var _ Expression = &LiteralExpression{}

type LiteralExpression struct {
	baseExpression
	Literal string
}

func (l *LiteralExpression) LiteralExpression() *LiteralExpression {
	return l
}

type Range struct {
	Low  rune
	High rune
}
