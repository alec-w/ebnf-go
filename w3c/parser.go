package w3c

import (
	"fmt"
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
		return p.parseExpression(true)
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
		case char == utf8.RuneError && width == 0:
			return expression, nil
		case p.isBasicLatinLetter(char):
			if p.isRuleEnd() {
				return expression, nil
			}
			fallthrough
		case char == '[' || char == '#' || char == '\'' || char == '"' || char == '(' || char == '|':
			if char == '|' {
				p.offset += width
			}
			next, err := p.parseExpression(false)
			if err != nil {
				return nil, err
			}
			if char == '|' {
				expression = p.parseExpressionsAsAlternates(expression, next)
			} else {
				expression = p.parseExpressionsAsList(expression, next)
			}
		default:
			return nil, fmt.Errorf("expected end of rule, another expression, or an expression alternate symbol")
		}
	}
	return expression, nil
}

func (p *Parser) parseSimpleExpression() (Expression, error) {
	p.skipWhitespace()
	var expression Expression
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	switch {
	case char == '[':
		fallthrough
	case char == '#':
		// character set, check next character to see if is a forbidden list
		return nil, fmt.Errorf("parsing character set expression not yet supported")
	case char == '"':
		fallthrough
	case char == '\'':
		// literal string
		expression = p.parseLiteralExpression()
	case p.isBasicLatinLetter(char):
		expression = &SymbolExpression{Symbol: p.parseSymbol()}
	default:
		// error
		return nil, fmt.Errorf("looking for start of expression but character at offset %d was not the start of an expression", p.offset)
	}
	return expression, nil
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
	} else {
		expression, err = p.parseSimpleExpression()
	}
	if err != nil {
		return nil, err
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

func (p *Parser) parseExpressionsAsList(first, second Expression) Expression {
	if alternateExpression := second.AlternateExpression(); alternateExpression != nil {
		// Get the first expression in second (the alternate) and make it the end of the list and then make that the first in the alternate
		// E.g. expression = A, next = alternate(B, C)
		// becomes alternate(list (A, B), C)
		// TODO handle parenthesis expressions
		list := p.expressionToList(first)
		toAppendToList := []Expression{alternateExpression.Expressions[0]}
		if firstFromAlternateAsList := toAppendToList[0].ListExpression(); firstFromAlternateAsList != nil {
			toAppendToList = firstFromAlternateAsList.Expressions
		}
		list.Expressions = append(list.Expressions, toAppendToList...)
		alternateExpression.Expressions = append([]Expression{list}, alternateExpression.Expressions[1:]...)
		return alternateExpression
	}
	if nextListExpression := second.ListExpression(); nextListExpression != nil {
		// TODO handle parenthesis expressions
		nextListExpression.Expressions = append(p.expressionToList(first).Expressions, nextListExpression.Expressions...)
		return nextListExpression
	}
	// TODO handle parenthesis expressions
	list := p.expressionToList(first)
	list.Expressions = append(list.Expressions, second)
	return list
}

func (p *Parser) expressionToList(expression Expression) *ListExpression {
	if expression := expression.ListExpression(); expression != nil {
		return expression
	}
	return &ListExpression{Expressions: []Expression{expression}}
}

func (p *Parser) parseExpressionsAsAlternates(first, second Expression) Expression {
	// Make the expression an alternate with the next expression the last item
	// E.g. expression = A, next = list(B, C)
	// becomes alternate(A, list(B, C))
	// TODO handle parenthesis expressions
	out := p.expressionToAlternates(first)
	toAppend := []Expression{second}
	if alternateExpression := second.AlternateExpression(); alternateExpression != nil {
		toAppend = alternateExpression.Expressions
	}
	out.Expressions = append(out.Expressions, toAppend...)
	return out
}

func (p *Parser) expressionToAlternates(expression Expression) *AlternateExpression {
	if expression := expression.AlternateExpression(); expression != nil {
		return expression
	}
	return &AlternateExpression{Expressions: []Expression{expression}}
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
