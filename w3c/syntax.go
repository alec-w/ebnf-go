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
	hasRepetitions() bool
}

type baseExpression struct {
	parenthesised bool
}

type Repetitions struct {
	Optional   bool `json:",omitempty"`
	OneOrMore  bool `json:",omitempty"`
	ZeroOrMore bool `json:",omitempty"`
}

func (r *Repetitions) setOptional(optional bool) {
	r.Optional = optional
}

func (r *Repetitions) setOneOrMore(oneOrMore bool) {
	r.OneOrMore = oneOrMore
}

func (r *Repetitions) setZeroOrMore(zeroOrMore bool) {
	r.ZeroOrMore = zeroOrMore
}

func (r *Repetitions) hasRepetitions() bool {
	return !repetitionsEmpty(*r)
}

func repetitionsEmpty(r Repetitions) bool {
	return !r.Optional && !r.OneOrMore && !r.ZeroOrMore
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

func (b *baseExpression) isParenthesised() bool {
	return b.parenthesised
}

func (b *baseExpression) setParenthesised(parenthesised bool) {
	b.parenthesised = parenthesised
}

var _ Expression = &ListExpression{}

type ListExpression struct {
	baseExpression
	Repetitions
	Expressions []Expression
}

func (a *ListExpression) Optional() bool {
	return a.Repetitions.Optional
}

func (a *ListExpression) OneOrMore() bool {
	return a.Repetitions.OneOrMore
}

func (a *ListExpression) ZeroOrMore() bool {
	return a.Repetitions.ZeroOrMore
}

func (a *ListExpression) ListExpression() *ListExpression {
	return a
}

func (a *ListExpression) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"list": a.Expressions,
		//		"Optional":    a.Optional(),
		//		"OneOrMore":   a.OneOrMore(),
		//		"ZeroOrMore":  a.ZeroOrMore(),
	})
}

var _ Expression = &AlternateExpression{}

type AlternateExpression struct {
	baseExpression
	Repetitions
	Expressions []Expression
}

func (l *AlternateExpression) Optional() bool {
	return l.Repetitions.Optional
}

func (l *AlternateExpression) OneOrMore() bool {
	return l.Repetitions.OneOrMore
}

func (l *AlternateExpression) ZeroOrMore() bool {
	return l.Repetitions.ZeroOrMore
}

func (l *AlternateExpression) AlternateExpression() *AlternateExpression {
	return l
}

func (l *AlternateExpression) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"alternate": l.Expressions,
		//		"Optional":    l.Optional(),
		//		"OneOrMore":   l.OneOrMore(),
		//		"ZeroOrMore":  l.ZeroOrMore(),
	})
}

var _ Expression = &ExceptionExpression{}

type ExceptionExpression struct {
	baseExpression
	Repetitions
	Match  Expression
	Except Expression
}

func (e *ExceptionExpression) Optional() bool {
	return e.Repetitions.Optional
}

func (e *ExceptionExpression) OneOrMore() bool {
	return e.Repetitions.OneOrMore
}

func (e *ExceptionExpression) ZeroOrMore() bool {
	return e.Repetitions.ZeroOrMore
}

func (e *ExceptionExpression) ExceptionExpression() *ExceptionExpression {
	return e
}

var _ Expression = &SymbolExpression{}

type SymbolExpression struct {
	baseExpression
	Repetitions
	Symbol string
}

func (s *SymbolExpression) Optional() bool {
	return s.Repetitions.Optional
}

func (s *SymbolExpression) OneOrMore() bool {
	return s.Repetitions.OneOrMore
}

func (s *SymbolExpression) ZeroOrMore() bool {
	return s.Repetitions.ZeroOrMore
}

func (e *SymbolExpression) SymbolExpression() *SymbolExpression {
	return e
}

var _ Expression = &CharacterSetExpression{}

type CharacterSetExpression struct {
	baseExpression
	Repetitions
	Enumerations []rune
	Ranges       []Range
	Forbidden    bool
}

func (c *CharacterSetExpression) Optional() bool {
	return c.Repetitions.Optional
}

func (c *CharacterSetExpression) OneOrMore() bool {
	return c.Repetitions.OneOrMore
}

func (c *CharacterSetExpression) ZeroOrMore() bool {
	return c.Repetitions.ZeroOrMore
}

func (c *CharacterSetExpression) CharacterSetExpression() *CharacterSetExpression {
	return c
}

var _ Expression = &LiteralExpression{}

type LiteralExpression struct {
	baseExpression
	Repetitions
	Literal string
}

func (l *LiteralExpression) Optional() bool {
	return l.Repetitions.Optional
}

func (l *LiteralExpression) OneOrMore() bool {
	return l.Repetitions.OneOrMore
}

func (l *LiteralExpression) ZeroOrMore() bool {
	return l.Repetitions.ZeroOrMore
}

func (l *LiteralExpression) LiteralExpression() *LiteralExpression {
	return l
}

type Range struct {
	Low  rune
	High rune
}
