package w3c

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type Parser struct {
	source string
	offset int
	line   int
}

func New() *Parser {
	return &Parser{}
}

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
		return Rule{}, fmt.Errorf("expected start of rule on line %d at total offset %d to begin with basic latin letter", p.line, p.offset)
	}
	rule := Rule{Symbol: p.parseSymbol(), Line: p.line}
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ':' {
		return Rule{}, fmt.Errorf("expected rule defining symbol on line %d at total offset %d to be ':=='", p.line, p.offset)
	}
	p.offset += width
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ':' {
		return Rule{}, fmt.Errorf("expected rule defining symbol on line %d at total offset %d to be '::='", p.line, p.offset)
	}
	p.offset += width
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '=' {
		return Rule{}, fmt.Errorf("expected rule defining symbol on line %d at total offset %d to be '::='", p.line, p.offset)
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
	for char, width := utf8.DecodeRuneInString(p.source[p.offset:]); p.isBasicLatinLetter(char); char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
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
			return nil, fmt.Errorf("expected closing parenthesis at end of parenthesised expression")
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
			return nil, fmt.Errorf("expected end of rule, another expression, or an expression alternate symbol")
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
		return nil, fmt.Errorf("looking for start of expression but character at offset %d was not the start of an expression", p.offset)
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
	for char, width = utf8.DecodeRuneInString(p.source[p.offset:]); char != terminalChar; char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
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
	for char, width = utf8.DecodeRuneInString(p.source[p.offset:]); char != ']'; char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
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
	for char, width := utf8.DecodeRuneInString(p.source[p.offset:]); char >= '0' && char <= '9'; char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
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
	for char, width := utf8.DecodeRuneInString(p.source[p.offset:]); char != ']'; char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
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
	// A or B have repetitions => list(A, B)
	// A or B is parethesised alternate => list(A, B)
	// A and B are both alternates => list(alt(A[:-1]), A[-1], B[0], alt(B[1:]))
	// A is alternate and B is not list => alt(A[:-1], list(A[-1], B))
	// A is alternate and B is list => alt(A[:-1], list(A[-1], B...))
	// B is alternate and A is not list => alt(list(A, B[0]), B[1:])
	// B is alternate and A is list => alt(list(A..., B[0]), B[1:])
	// A is list and B is not list => list(A..., B)
	// A is not list and B is list => list(A, B...)
	// A is list and B is list => list(A..., B...)
	// else => list(A, B)
	aAsAlternate := a.AlternateExpression()
	bAsAlternate := b.AlternateExpression()
	aAsList := a.ListExpression()
	bAsList := b.ListExpression()
	if (a.hasRepetitions() || b.hasRepetitions()) ||
		(aAsAlternate != nil && a.isParenthesised()) ||
		(bAsAlternate != nil && b.isParenthesised()) ||
		((aAsList == nil && bAsList == nil) && (aAsAlternate == nil && bAsAlternate == nil)) {
		return &ListExpression{Expressions: []Expression{a, b}}
	}
	if aAsAlternate != nil && bAsAlternate != nil {
		middle := []Expression{aAsAlternate.Expressions[len(aAsAlternate.Expressions)-1], bAsAlternate.Expressions[0]}
		aAsAlternate.Expressions = aAsAlternate.Expressions[:len(aAsAlternate.Expressions)-1]
		bAsAlternate.Expressions = bAsAlternate.Expressions[1:]
		return &AlternateExpression{Expressions: append(
			append(
				aAsAlternate.Expressions,
				middle...,
			),
			bAsAlternate.Expressions...,
		),
		}
	}
	if aAsAlternate != nil {
		listExpression := &ListExpression{Expressions: []Expression{aAsAlternate.Expressions[len(aAsAlternate.Expressions)-1]}}
		if first := listExpression.Expressions[0].ListExpression(); first != nil {
			listExpression = first
		}
		aAsAlternate.Expressions = aAsAlternate.Expressions[:len(aAsAlternate.Expressions)-1]
		if bAsList != nil {
			listExpression.Expressions = append(listExpression.Expressions, bAsList.Expressions...)
		} else {
			listExpression.Expressions = append(listExpression.Expressions, b)
		}
		return &AlternateExpression{Expressions: append(aAsAlternate.Expressions, listExpression)}
	}
	if bAsAlternate != nil {
		listExpression := &ListExpression{Expressions: []Expression{bAsAlternate.Expressions[0]}}
		if first := listExpression.Expressions[0].ListExpression(); first != nil {
			listExpression = first
		}
		bAsAlternate.Expressions = bAsAlternate.Expressions[1:]
		if aAsList != nil {
			listExpression.Expressions = append(aAsList.Expressions, listExpression.Expressions...)
		} else {
			listExpression.Expressions = append([]Expression{a}, listExpression.Expressions...)
		}
		return &AlternateExpression{Expressions: append([]Expression{listExpression}, bAsAlternate.Expressions...)}
	}
	expressions := []Expression{a}
	if aAsList != nil {
		expressions = aAsList.Expressions
	}
	expressions = append(expressions, b)
	if bAsList != nil {
		expressions = expressions[:len(expressions)-1]
		expressions = append(expressions, bAsList.Expressions...)
	}
	return &ListExpression{Expressions: expressions}
}

func (p *Parser) parseExpressionsAsAlternates(a, b Expression) Expression {
	// A | B
	// if A or B has repetitions => alt(A, B)
	// if A is alternate => first_terms = A... else A
	// if B is alternate => second_terms = B... else B
	// => alt(first_terms..., second_terms...)
	if a.hasRepetitions() || b.hasRepetitions() {
		return &AlternateExpression{Expressions: []Expression{a, b}}
	}
	firstTerms := []Expression{a}
	if aAsAlternate := a.AlternateExpression(); aAsAlternate != nil {
		firstTerms = aAsAlternate.Expressions
	}
	secondTerms := []Expression{b}
	if bAsAlternate := b.AlternateExpression(); bAsAlternate != nil {
		secondTerms = bAsAlternate.Expressions
	}
	return &AlternateExpression{Expressions: append(firstTerms, secondTerms...)}
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
	p.parseSymbol()
	p.skipWhitespace()
	potentialDefiningSymbolOffset := p.offset
	p.offset = startOffset
	return p.source[potentialDefiningSymbolOffset:potentialDefiningSymbolOffset+3] == "::="
}

func (p *Parser) isBasicLatinLetter(char rune) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

func (p *Parser) skipWhitespace() {
	for char, width := utf8.DecodeRuneInString(p.source[p.offset:]); unicode.IsSpace(char); char, width = utf8.DecodeRuneInString(p.source[p.offset:]) {
		p.offset += width
		if char == '\n' {
			p.line++
		}
	}
}
