package w3c

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Parser parses an EBNF grammar into a Syntax.
type Parser struct {
	source string
	offset int
	line   int
}

// New instantiates a Parser.
func New() *Parser {
	return &Parser{}
}

// Parse parses the given EBNF grammar into a Syntax representation.
func (p *Parser) Parse(source string) (Syntax, error) {
	p.source = source
	p.offset = 0
	p.line = 1
	syntax, err := p.parseSyntax()

	return syntax, err
}

func (p *Parser) parseSyntax() (Syntax, error) {
	var syntax Syntax
	for p.skipWhitespace(); p.source[p.offset:] != ""; p.skipWhitespace() {
		rule, err := p.parseRule()
		if err != nil {
			return Syntax{}, err
		}
		syntax.Rules = append(syntax.Rules, rule)
	}

	return syntax, nil
}

func (p *Parser) parseRule() (Rule, error) {
	p.skipWhitespace()
	if char, _ := utf8.DecodeRuneInString(p.source[p.offset:]); !p.isBasicLatinLetter(char) {
		return Rule{}, fmt.Errorf(
			"expected start of rule on line %d at total offset %d to begin with basic latin letter",
			p.line,
			p.offset,
		)
	}
	rule := Rule{Symbol: p.parseSymbol(), Line: p.line}
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ':' {
		return Rule{}, fmt.Errorf(
			"expected rule defining symbol on line %d at total offset %d to be ':=='",
			p.line,
			p.offset,
		)
	}
	p.offset += width
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ':' {
		return Rule{}, fmt.Errorf(
			"expected rule defining symbol on line %d at total offset %d to be '::='",
			p.line,
			p.offset,
		)
	}
	p.offset += width
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '=' {
		return Rule{}, fmt.Errorf(
			"expected rule defining symbol on line %d at total offset %d to be '::='",
			p.line,
			p.offset,
		)
	}
	p.offset += width
	expression, err := p.parseExpression(false)
	if err != nil {
		return Rule{}, err
	}
	rule.Expression = expression

	return rule, nil
}

func (p *Parser) parseSymbol() string {
	var symbol []rune
	for char, width := p.next(); p.isBasicLatinLetter(char); char, width = p.next() {
		symbol = append(symbol, char)
		p.offset += width
	}

	return string(symbol)
}

/*
Expressions are parsed recursively for the shortest distance, then combined together
A | (B | C) D
Is parsed as:
1st call - expression = A, next = (B | C) D - returns A | (B | C) D
2nd call (invoked from 1st) - expression = (B | C), next = D - returns (B | C) D
3rd call (invoked from 2nd) - expression = B, next = C - returns B | C
4th call (invoked from 3rd) - expression = C, next = nil - returns C
5th call (invoked from 2nd) - expression = D, next = nil - returns D
*/

func (p *Parser) parseExpression(_ bool) (Expression, error) {
	var expression Expression
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '(' {
		p.offset += width
		var err error
		expression, err = p.parseExpression(true)
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != ')' {
			return nil, fmt.Errorf(
				"expected closing parenthesis at end of parenthesised expression",
			)
		}
		p.offset += width
		expression.setParenthesised(true)
	} else {
		var err error
		expression, err = p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
	}
	expression = p.parseExpressionRepetitions(expression)
	expression, err := p.parseExpressionException(expression)
	if err != nil {
		return nil, err
	}
	for !p.isRuleEnd() {
		p.skipWhitespace()
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
		switch {
		case p.isBasicLatinLetter(char) || char == '[' || char == '#' || char == '\'' || char == '"' || char == '(':
			next, err := p.parseExpression(false)
			if err != nil {
				return nil, err
			}
			expression = p.parseExpressionsAsList(expression, next)
		case char == '|':
			p.offset += width
			next, err := p.parseExpression(false)
			if err != nil {
				return nil, err
			}
			expression = p.parseExpressionsAsAlternates(expression, next)
		case char == ')':
			return expression, nil
		default:
			return nil, fmt.Errorf(
				"expected end of rule, another expression, or an expression alternate symbol",
			)
		}
	}

	return expression, nil
}

func (p *Parser) parseSimpleExpression() (Expression, error) {
	p.skipWhitespace()
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	switch {
	case char == '[':
		fallthrough
	case char == '#':
		// character set, check next character to see if is a forbidden list
		return p.parseCharacterSetExpression()
	case char == '"':
		fallthrough
	case char == '\'':
		// literal string
		return p.parseLiteralExpression(), nil
	case p.isBasicLatinLetter(char):
		return &SymbolExpression{Symbol: p.parseSymbol()}, nil
	default:
		// error
		return nil, fmt.Errorf(
			"looking for start of expression but character at offset %d was not the start of an expression",
			p.offset,
		)
	}
}

func (p *Parser) parseExpressionRepetitions(expression Expression) Expression {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	switch char {
	case '?':
		expression.setOptional(true)
		p.offset += width
	case '*':
		expression.setZeroOrMore(true)
		p.offset += width
	case '+':
		expression.setOneOrMore(true)
		p.offset += width
	default:
		// No repetitions
	}

	return expression
}

func (p *Parser) parseExpressionException(expression Expression) (Expression, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '-' {
		return expression, nil
	}
	p.offset += width
	exceptExpression := &ExceptionExpression{Match: expression}
	var err error
	p.skipWhitespace()
	char, _ = utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '(' {
		expression, err = p.parseExpression(false)
		if err != nil {
			return nil, err
		}
	} else {
		expression, err = p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
		expression = p.parseExpressionRepetitions(expression)
	}
	exceptExpression.Except = expression

	return exceptExpression, nil
}

func (p *Parser) parseLiteralExpression() *LiteralExpression {
	terminalChar, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	expression := &LiteralExpression{}
	var char rune
	for char, width = p.next(); char != terminalChar; char, width = p.next() {
		expression.Literal += string(char)
		p.offset += width
	}
	p.offset += width

	return expression
}

func (p *Parser) parseCharacterSetExpression() (*CharacterSetExpression, error) {
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '#' {
		char, err := p.parseHexCharacter()
		if err != nil {
			return nil, err
		}

		return &CharacterSetExpression{Enumerations: []rune{char}}, nil
	}
	p.offset += width
	expression := &CharacterSetExpression{}
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '^' {
		expression.Forbidden = true
		p.offset += width
	}
	for char, width = p.next(); char != ']'; char, width = p.next() {
		if char == '#' {
			var err error
			char, err = p.parseHexCharacter()
			if err != nil {
				return nil, err
			}
		} else {
			p.offset += width
		}
		first := char
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
		if char == '-' {
			p.offset += width
			char, width = utf8.DecodeRuneInString(p.source[p.offset:])
			if char == '#' {
				var err error
				char, err = p.parseHexCharacter()
				if err != nil {
					return nil, err
				}
			} else {
				p.offset += width
			}
			expression.Ranges = append(expression.Ranges, Range{Low: first, High: char})
		} else {
			enumeration, err := p.parseCharacterSetEnumeration()
			if err != nil {
				return nil, err
			}
			expression.Enumerations = append(expression.Enumerations, first)
			expression.Enumerations = append(expression.Enumerations, enumeration...)
		}
	}
	p.offset += width

	return expression, nil
}

func (p *Parser) parseHexCharacter() (rune, error) {
	_, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != 'x' {
		return 0, fmt.Errorf("did not get x after # in hex character")
	}
	p.offset += width
	var chars []rune
	for char, width := p.next(); char >= '0' && char <= '9'; char, width = p.next() {
		p.offset += width
		chars = append(chars, char)
	}
	intVal, err := strconv.ParseUint(string(chars), 16, 32)
	if err != nil {
		return 0, err
	}

	return rune(intVal), nil
}

func (p *Parser) parseCharacterSetEnumeration() ([]rune, error) {
	var chars []rune
	var offsets []int
	for char, width := p.next(); char != ']'; char, width = p.next() {
		offsets = append(offsets, p.offset)
		if char == '#' {
			var err error
			char, err = p.parseHexCharacter()
			if err != nil {
				return nil, err
			}
		} else {
			p.offset += width
		}
		chars = append(chars, char)
		if char == '-' {
			p.offset = offsets[len(offsets)-2]
			chars = chars[:len(chars)-2]

			break
		}
	}

	return chars, nil
}

func (p *Parser) parseExpressionsAsList(a, b Expression) Expression {
	// A B
	// B is simple expression
	// (A1 | A2 | A3) B => list(A, B)
	// (A1 A2 A3)? B => list(A, B)
	// A1 A2 A3 B => list(A..., B)
	// A1 | A2 | A3 B => alt(A[:-1]..., listJoin(A[-1], B))
	// A B => list(A, B)
	// B is parenthesised alternate
	// (A1 | A2 | A3) (B1 | B2 | B3) => list(A, B)
	// (A1 A2 A3)? (B1 | B2 | B3) => list(A, B)
	// A1 A2 A3 (B1 | B2 | B3) => list(A... B)
	// A1 | A2 | A3 (B1 | B2 | B3) => alt(A[:-1]..., listJoin(A[-1], B))
	// A (B1 | B2 | B3) => list(A, B)
	// B has repetitions
	// (A1 | A2 | A3) (B1 B2 B3)? => list(A, B)
	// (A1 A2 A3)? (B1 B2 B3)? => list(A, B)
	// A1 A2 A3 (B1 B2 B3)? => list(A..., B)
	// A1 | A2 | A3 (B1 B2 B3)? => alt(A[:-1]..., listJoin(A[-1], B))
	// A (B1 B2 B3)? => list(A, B)
	// B is list without repetitions
	// (A1 | A2 | A3) B1 B2 B3 => list(A, B...)
	// (A1 A2 A3)? B1 B2 B3 => list(A, B...)
	// A1 A2 A3 B1 B2 B3 => list(A..., B...)
	// A1 | A2 | A3 B1 B2 B3 => alt(A[:-1]..., listJoin(A[-1], B))
	// A B1 B2 B3 => list(A, B...)
	// B is alternate without parentheses or repetitions
	// (A1 | A2 | A3) B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// (A1 A2 A3)? B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A1 A2 A3 B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A1 | A2 | A3 B1 | B2 | B3 => alt(A[:-1]..., listJoin(A[-1], B[0]), B[1:]...)
	//
	// Reordered
	// Set 1
	// (A1 | A2 | A3) B => list(A, B)
	// (A1 A2 A3)? B => list(A, B)
	// A B => list(A, B)
	// (A1 | A2 | A3) (B1 | B2 | B3) => list(A, B)
	// (A1 A2 A3)? (B1 | B2 | B3) => list(A, B)
	// A (B1 | B2 | B3) => list(A, B)
	// (A1 | A2 | A3) (B1 B2 B3)? => list(A, B)
	// (A1 A2 A3)? (B1 B2 B3)? => list(A, B)
	// A (B1 B2 B3)? => list(A, B)
	// Set 2a
	// A1 A2 A3 B => list(A..., B)
	// A1 A2 A3 (B1 | B2 | B3) => list(A... B)
	// A1 A2 A3 (B1 B2 B3)? => list(A..., B)
	// Set 2b
	// (A1 | A2 | A3) B1 B2 B3 => list(A, B...)
	// (A1 A2 A3)? B1 B2 B3 => list(A, B...)
	// A B1 B2 B3 => list(A, B...)
	// Set 2c
	// A1 A2 A3 B1 B2 B3 => list(A..., B...)
	// Set 3a
	// A1 | A2 | A3 B => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 (B1 | B2 | B3) => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 (B1 B2 B3)? => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 B1 B2 B3 => alt(A[:-1]..., listJoin(A[-1], B))
	// Set 3b
	// (A1 | A2 | A3) B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// (A1 A2 A3)? B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A1 A2 A3 B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// Set 3c
	// A1 | A2 | A3 B1 | B2 | B3 => alt(A[:-1]..., listJoin(A[-1], B[0]), B[1:]...)
	aAsList := a.ListExpression()
	bAsList := b.ListExpression()
	aAsAlternate := a.AlternateExpression()
	bAsAlternate := b.AlternateExpression()
	aIsSimpleExpression := aAsAlternate == nil && aAsList == nil
	bIsSimpleExpression := bAsAlternate == nil && bAsList == nil
	// Set 1
	// (A1 | A2 | A3) B => list(A, B)
	// (A1 A2 A3)? B => list(A, B)
	// A B => list(A, B)
	// (A1 | A2 | A3) (B1 | B2 | B3) => list(A, B)
	// (A1 A2 A3)? (B1 | B2 | B3) => list(A, B)
	// A (B1 | B2 | B3) => list(A, B)
	// (A1 | A2 | A3) (B1 B2 B3)? => list(A, B)
	// (A1 A2 A3)? (B1 B2 B3)? => list(A, B)
	// A (B1 B2 B3)? => list(A, B)
	if ((aAsAlternate != nil && a.isParenthesised()) || a.hasRepetitions() || (aIsSimpleExpression)) &&
		((bAsAlternate != nil && b.isParenthesised()) || b.hasRepetitions() || (bIsSimpleExpression)) {
		return &ListExpression{Expressions: []Expression{a, b}}
	}
	// Set 2a
	// A1 A2 A3 B => list(A..., B)
	// A1 A2 A3 (B1 | B2 | B3) => list(A... B)
	// A1 A2 A3 (B1 B2 B3)? => list(A..., B)
	// Set 2b
	// (A1 | A2 | A3) B1 B2 B3 => list(A, B...)
	// (A1 A2 A3)? B1 B2 B3 => list(A, B...)
	// A B1 B2 B3 => list(A, B...)
	// Set 2c
	// A1 A2 A3 B1 B2 B3 => list(A..., B...)
	if (aAsList != nil && ((bAsAlternate != nil && b.isParenthesised()) || b.hasRepetitions() || (bIsSimpleExpression))) ||
		(bAsList != nil && ((aAsAlternate != nil && a.isParenthesised()) || a.hasRepetitions() || (aIsSimpleExpression))) {
		var expressions []Expression
		if aAsList != nil && !a.hasRepetitions() {
			expressions = append(expressions, aAsList.Expressions...)
		} else {
			expressions = append(expressions, a)
		}
		if bAsList != nil && !b.hasRepetitions() {
			expressions = append(expressions, bAsList.Expressions...)
		} else {
			expressions = append(expressions, b)
		}

		return &ListExpression{Expressions: expressions}
	}
	// Set 3a
	// A1 | A2 | A3 B => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 (B1 | B2 | B3) => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 (B1 B2 B3)? => alt(A[:-1]..., listJoin(A[-1], B))
	// A1 | A2 | A3 B1 B2 B3 => alt(A[:-1]..., listJoin(A[-1], B))
	// Set 3b
	// (A1 | A2 | A3) B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// (A1 A2 A3)? B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A1 A2 A3 B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// A B1 | B2 | B3 => alt(listJoin(A, B[0]), B[1:]...)
	// Set 3c
	// A1 | A2 | A3 B1 | B2 | B3 => alt(A[:-1]..., listJoin(A[-1], B[0]), B[1:]...)
	if aAsAlternate != nil && bAsAlternate != nil {
		return &AlternateExpression{Expressions: append(
			append(
				aAsAlternate.Expressions[:len(aAsAlternate.Expressions)-1],
				p.parseExpressionsAsList(
					aAsAlternate.Expressions[len(aAsAlternate.Expressions)-1],
					bAsAlternate.Expressions[0],
				),
			),
			bAsAlternate.Expressions[1:]...,
		)}
	} else if aAsAlternate != nil {
		return &AlternateExpression{Expressions: append(
			aAsAlternate.Expressions[:len(aAsAlternate.Expressions)-1],
			p.parseExpressionsAsList(aAsAlternate.Expressions[len(aAsAlternate.Expressions)-1], b),
		)}
	}

	return &AlternateExpression{Expressions: append(
		[]Expression{p.parseExpressionsAsList(a, bAsAlternate.Expressions[0])},
		bAsAlternate.Expressions[1:]...,
	)}
}

func (p *Parser) parseExpressionsAsAlternates(a, b Expression) Expression {
	aAsAlternate := a.AlternateExpression()
	bAsAlternate := b.AlternateExpression()
	var expressions []Expression
	if aAsAlternate != nil && !a.hasRepetitions() {
		expressions = append(expressions, aAsAlternate.Expressions...)
	} else {
		expressions = append(expressions, a)
	}
	if bAsAlternate != nil && !b.hasRepetitions() {
		expressions = append(expressions, bAsAlternate.Expressions...)
	} else {
		expressions = append(expressions, b)
	}

	return &AlternateExpression{Expressions: expressions}
}

func (p *Parser) isRuleEnd() bool {
	p.skipWhitespace()
	if p.source[p.offset:] == "" {
		return true
	}
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	if !p.isBasicLatinLetter(char) {
		return false
	}
	startOffset := p.offset
	startLine := p.line
	p.parseSymbol()
	p.skipWhitespace()
	potentialDefiningSymbolOffset := p.offset
	p.offset = startOffset
	p.line = startLine
	if potentialDefiningSymbolOffset+3 >= len(p.source) {
		return false
	}

	return p.source[potentialDefiningSymbolOffset:potentialDefiningSymbolOffset+3] == "::="
}

func (p *Parser) isBasicLatinLetter(char rune) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

func (p *Parser) skipWhitespace() {
	for char, width := p.next(); unicode.IsSpace(char); char, width = p.next() {
		p.offset += width
		if char == '\n' {
			p.line++
		}
	}
}

func (p *Parser) next() (rune, int) {
	return utf8.DecodeRuneInString(p.source[p.offset:])
}
