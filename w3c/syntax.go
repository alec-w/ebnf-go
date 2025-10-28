package w3c

import "encoding/json"

// Syntax is a top level EBNF grammar.
type Syntax struct {
	Rules []Rule `json:"rules"`
}

// Rule is a single rule from an EBNF grammar.
type Rule struct {
	Line       int        `json:"line"`
	Symbol     string     `json:"symbol"`
	Expression Expression `json:"expression"`
}

// Expression is fulfilled by every expression type.
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

// baseExpression is used to give every expression the option of being parenthesised.
//
// Whether an expression is parenthesised does not need to be exposed externally, this is only used when parsing a
// grammar.
type baseExpression struct {
	parenthesised bool
}

// Repetitions records whether an expression is repeated and in what fashion, a maximum of one field in this struct
// should be true for an expression.
type Repetitions struct {
	Optional   bool `json:"optional,omitempty"`
	OneOrMore  bool `json:"oneOrMore,omitempty"`
	ZeroOrMore bool `json:"zeroOrMore,omitempty"`
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

// ListExpression fulfils the Expression interface.
func (b *baseExpression) ListExpression() *ListExpression {
	return nil
}

// AlternateExpression fulfils the Expression interface.
func (b *baseExpression) AlternateExpression() *AlternateExpression {
	return nil
}

// ExceptionExpression fulfils the Expression interface.
func (b *baseExpression) ExceptionExpression() *ExceptionExpression {
	return nil
}

// SymbolExpression fulfils the Expression interface.
func (b *baseExpression) SymbolExpression() *SymbolExpression {
	return nil
}

// CharacterSetExpression fulfils the Expression interface.
func (b *baseExpression) CharacterSetExpression() *CharacterSetExpression {
	return nil
}

// LiteralExpression fulfils the Expression interface.
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

// ListExpression represents a list of expressions concatenated together to form a larger expression.
type ListExpression struct {
	baseExpression
	Repetitions
	Expressions []Expression
}

// Optional fulfils the Expression interface.
func (l *ListExpression) Optional() bool {
	return l.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (l *ListExpression) OneOrMore() bool {
	return l.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (l *ListExpression) ZeroOrMore() bool {
	return l.Repetitions.ZeroOrMore
}

// ListExpression exposes the underlying ListExpression.
func (l *ListExpression) ListExpression() *ListExpression {
	return l
}

// MarshalJSON fulfils the json.Marshaller interface.
func (l *ListExpression) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"list": l.Expressions,
	}
	if l.Optional() {
		out["optional"] = true
	}
	if l.OneOrMore() {
		out["oneOrMore"] = true
	}
	if l.ZeroOrMore() {
		out["zeroOrMore"] = true
	}

	return json.Marshal(out)
}

var _ Expression = &AlternateExpression{}

// AlternateExpression represents an expression that is fulfilled by a single one of a group of expressions.
type AlternateExpression struct {
	baseExpression
	Repetitions
	Expressions []Expression
}

// Optional fulfils the Expression interface.
func (a *AlternateExpression) Optional() bool {
	return a.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (a *AlternateExpression) OneOrMore() bool {
	return a.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (a *AlternateExpression) ZeroOrMore() bool {
	return a.Repetitions.ZeroOrMore
}

// AlternateExpression exposes the underlying AlternateExpression.
func (a *AlternateExpression) AlternateExpression() *AlternateExpression {
	return a
}

// MarshalJSON fulfils the json.Marshaller interface.
func (a *AlternateExpression) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"alternate": a.Expressions,
	}
	if a.Optional() {
		out["optional"] = true
	}
	if a.OneOrMore() {
		out["oneOrMore"] = true
	}
	if a.ZeroOrMore() {
		out["zeroOrMore"] = true
	}

	return json.Marshal(out)
}

var _ Expression = &ExceptionExpression{}

// ExceptionExpression represents an expression that is fulfilled by an expression matching the Match expression but not
// matching the Except expression.
type ExceptionExpression struct {
	baseExpression
	Repetitions
	Match  Expression `json:"match"`
	Except Expression `json:"except"`
}

// Optional fulfils the Expression interface.
func (e *ExceptionExpression) Optional() bool {
	return e.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (e *ExceptionExpression) OneOrMore() bool {
	return e.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (e *ExceptionExpression) ZeroOrMore() bool {
	return e.Repetitions.ZeroOrMore
}

// ExceptionExpression exposes the underlying ExceptionExpression.
func (e *ExceptionExpression) ExceptionExpression() *ExceptionExpression {
	return e
}

var _ Expression = &SymbolExpression{}

// SymbolExpression represents an expression that references a symbol (a rule's expression).
type SymbolExpression struct {
	baseExpression
	Repetitions
	Symbol string `json:"symbol"`
}

// Optional fulfils the Expression interface.
func (s *SymbolExpression) Optional() bool {
	return s.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (s *SymbolExpression) OneOrMore() bool {
	return s.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (s *SymbolExpression) ZeroOrMore() bool {
	return s.Repetitions.ZeroOrMore
}

// SymbolExpression exposes the underlying SymbolExpression.
func (s *SymbolExpression) SymbolExpression() *SymbolExpression {
	return s
}

var _ Expression = &CharacterSetExpression{}

// CharacterSetExpression represents an expression that is fulfilled by a character in any of a set of UTF8 ranges or
// direct character enumerations. If Forbidden is true then the character must not fall within any of the
// ranges/enumerations.
type CharacterSetExpression struct {
	baseExpression
	Repetitions
	Enumerations []rune  `json:"enumerations,omitempty"`
	Ranges       []Range `json:"ranges,omitempty"`
	Forbidden    bool    `json:"forbidden,omitempty"`
}

// Optional fulfils the Expression interface.
func (c *CharacterSetExpression) Optional() bool {
	return c.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (c *CharacterSetExpression) OneOrMore() bool {
	return c.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (c *CharacterSetExpression) ZeroOrMore() bool {
	return c.Repetitions.ZeroOrMore
}

// CharacterSetExpression exposes the underlying CharacterSetExpression.
func (c *CharacterSetExpression) CharacterSetExpression() *CharacterSetExpression {
	return c
}

var _ Expression = &LiteralExpression{}

// LiteralExpression represents an expression fulfilled by a literal sequence of characters.
type LiteralExpression struct {
	baseExpression
	Repetitions
	Literal string `json:"literal"`
}

// Optional fulfils the Expression interface.
func (l *LiteralExpression) Optional() bool {
	return l.Repetitions.Optional
}

// OneOrMore fulfils the Expression interface.
func (l *LiteralExpression) OneOrMore() bool {
	return l.Repetitions.OneOrMore
}

// ZeroOrMore fulfils the Expression interface.
func (l *LiteralExpression) ZeroOrMore() bool {
	return l.Repetitions.ZeroOrMore
}

// LiteralExpression exposes the underlying LiteralExpression.
func (l *LiteralExpression) LiteralExpression() *LiteralExpression {
	return l
}

// Range represents a UTF8 character range.
type Range struct {
	Low  rune `json:"low"`
	High rune `json:"high"`
}
