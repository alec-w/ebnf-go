package ebnf

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type Parser struct {
	source string
	offset int
}

func New() Parser {
	return Parser{}
}

func (p *Parser) Parse(source string) (Syntax, error) {
	p.source = source
	p.offset = 0
	return p.parseSyntax()
}

func (p *Parser) parseSyntax() (Syntax, error) {
	var syntax Syntax
	rule, err := p.parseRule()
	if err != nil {
		return Syntax{}, err
	}
	syntax.Rules = append(syntax.Rules, rule)
	p.skipWhitespace()
	for p.source[p.offset:] != "" {
		rule, err := p.parseRule()
		if err != nil {
			return Syntax{}, err
		}
		syntax.Rules = append(syntax.Rules, rule)
		p.skipWhitespace()
	}
	return syntax, nil
}

func (p *Parser) parseRule() (Rule, error) {
	// Remove leading whitespace
	p.skipWhitespace()
	// Look for start of meta identifier (letter)
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	if !unicode.IsLetter(char) {
		return Rule{}, fmt.Errorf(
			"rule meta identifier does not start with letter, starts with %q",
			char,
		)
	}
	// Parse the meta identifier
	metaIdentifier, err := p.parseMetaIdentifier()
	if err != nil {
		return Rule{}, err
	}
	rule := Rule{MetaIdentifier: metaIdentifier}
	// Remove leading whitespace
	p.skipWhitespace()
	// Look for "=" character and remove it
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '=' {
		return Rule{}, fmt.Errorf(
			"next non-whitespace character after rule meta identifier should be %q but character at offset %d was not",
			'=',
			p.offset,
		)
	}
	p.offset += width
	// Parse a definitions list
	defintitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return Rule{}, err
	}
	rule.Definitions = defintitionsList
	// Look for terminating character
	p.skipWhitespace()
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ';' && char != '.' {
		return Rule{}, fmt.Errorf(
			"no terminator symbol (one of %q or %q) found at end of syntax rule at offset %d",
			'.',
			';',
			p.offset,
		)
	}
	p.offset += width
	return rule, nil
}

func (p *Parser) parseMetaIdentifier() (string, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if !unicode.IsLetter(char) {
		return "", fmt.Errorf(
			"parsing meta identifier at offset %d but first character was not letter",
			p.offset,
		)
	}
	startOffset := p.offset
	for unicode.IsLetter(char) || unicode.IsDigit(char) {
		p.offset += width
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	return p.source[startOffset:p.offset], nil
}

func (p *Parser) parseDefinitionsList() (DefinitionsList, error) {
	var definitionsList DefinitionsList
	definition, err := p.parseDefinition()
	if err != nil {
		return nil, err
	}
	definitionsList = append(definitionsList, definition)
	p.skipWhitespace()
	next, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for next == '|' || next == '/' || next == '!' {
		if next == '/' {
			if next2, _ := utf8.DecodeRuneInString(p.source[p.offset+width:]); next2 == ')' {
				// "/)" is the end of an optional sequence
				break
			}
		}
		p.offset += width
		definition, err = p.parseDefinition()
		if err != nil {
			return nil, err
		}
		definitionsList = append(definitionsList, definition)
		p.skipWhitespace()
		next, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	return definitionsList, nil
}

func (p *Parser) parseDefinition() (Definition, error) {
	var definition Definition
	term, err := p.parseTerm()
	if err != nil {
		return Definition{}, err
	}
	definition.Terms = append(definition.Terms, term)
	p.skipWhitespace()
	next, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for next == ',' {
		p.offset += width
		term, err = p.parseTerm()
		if err != nil {
			return Definition{}, err
		}
		definition.Terms = append(definition.Terms, term)
		next, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	return definition, nil
}

func (p *Parser) parseTerm() (Term, error) {
	var term Term
	factor, err := p.parseFactor()
	if err != nil {
		return Term{}, err
	}
	term.Factor = factor
	p.skipWhitespace()
	next, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if next == '-' {
		p.offset += width
		p.skipWhitespace()
		exception, err := p.parseFactor()
		if err != nil {
			return Term{}, err
		}
		term.Exception = exception
	}
	return term, nil
}

func (p *Parser) parseFactor() (Factor, error) {
	factor := Factor{Repetitions: -1}
	p.skipWhitespace()
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	if unicode.IsDigit(char) {
		integer, err := p.parseInteger()
		if err != nil {
			return Factor{}, err
		}
		factor.Repetitions = integer
		p.skipWhitespace()
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != '*' {
			return Factor{}, fmt.Errorf(
				"parsing factor had integer for repetitions but non-whitespace character after was not %q at offset %d",
				'*',
				p.offset,
			)
		}
		p.offset += width
	}
	primary, err := p.parsePrimary()
	if err != nil {
		return Factor{}, err
	}
	factor.Primary = primary
	return factor, nil
}

func (p *Parser) parseInteger() (int, error) {
	p.skipWhitespace()
	startOffset := p.offset
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if !unicode.IsDigit(char) {
		return 0, fmt.Errorf(
			"parsing integer but first character was not digit at offset %d",
			p.offset,
		)
	}
	p.offset += width
	for ; unicode.IsDigit(char); p.offset += width {
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	parsedInt, err := strconv.Atoi(p.source[startOffset : p.offset-width])
	if err != nil {
		return 0, fmt.Errorf(
			"parsed integer could not be converted to integer type at offset %d",
			startOffset,
		)
	}
	return parsedInt, nil
}

func (p *Parser) parsePrimary() (Primary, error) {
	// Remove leading whitespace
	p.skipWhitespace()
	primary := Primary{}
	var err error
	var width int
	// Parse one of:
	// optional sequence
	handleOptionalSequence := func() error {
		var optionalSequence DefinitionsList
		optionalSequence, err = p.parseOptionalSequence()
		if err != nil {
			return err
		}
		width = 0
		primary.OptionalSequence = optionalSequence
		return nil
	}
	// repeated sequence
	handleRepeatedSequence := func() error {
		var repeatedSequence DefinitionsList
		repeatedSequence, err = p.parseRepeatedSequence()
		if err != nil {
			return err
		}
		width = 0
		primary.RepeatedSequence = repeatedSequence
		return nil
	}
	// special sequence
	// grouped sequence
	// meta identifier
	// terminal
	// empty
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	switch {
	case char == '[':
		if err := handleOptionalSequence(); err != nil {
			return Primary{}, err
		}
	case char == '{':
		if err := handleRepeatedSequence(); err != nil {
			return Primary{}, err
		}
	case char == '?':
		var specialSequence string
		specialSequence, err = p.parseSpecialSequence()
		if err != nil {
			return Primary{}, err
		}
		width = 0
		primary.SpecialSequence = specialSequence
	case char == '(':
		next, _ := utf8.DecodeRuneInString(p.source[p.offset+width:])
		switch next {
		case '/':
			if err := handleOptionalSequence(); err != nil {
				return Primary{}, err
			}
		case ':':
			if err := handleRepeatedSequence(); err != nil {
				return Primary{}, err
			}
		default:
			var groupedSequence DefinitionsList
			groupedSequence, err = p.parseGroupedSequence()
			if err != nil {
				return Primary{}, err
			}
			width = 0
			primary.GroupedSequence = groupedSequence
		}
	case unicode.IsLetter(char):
		var metaIdentifier string
		metaIdentifier, err = p.parseMetaIdentifier()
		if err != nil {
			return Primary{}, err
		}
		width = 0
		primary.MetaIdentifier = metaIdentifier
	case char == '\'':
		fallthrough
	case char == '"':
		var terminal string
		terminal, err = p.parseTerminal()
		if err != nil {
			return Primary{}, err
		}
		width = 0
		primary.Terminal = terminal
	default:
		width = 0
		primary.Empty = true
	}
	p.offset += width
	return primary, nil
}

func (p *Parser) parseOptionalSequence() (DefinitionsList, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	if char != '[' {
		if char != '(' {
			return nil, fmt.Errorf(
				"parsing optional sequence at offset %d but did not start with %q or %q",
				p.offset,
				'[',
				"(/",
			)
		}
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != '/' {
			return nil, fmt.Errorf(
				"parsing optional sequence at offset %d but did not start with %q or %q",
				p.offset,
				'[',
				"(/",
			)
		}
		p.offset += width
	}
	definitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	if char != ']' {
		if char != '/' {
			return nil, fmt.Errorf(
				"parsing optional sequence at offset %d but did not end with %q or %q",
				p.offset,
				']',
				"/)",
			)
		}
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != ')' {
			return nil, fmt.Errorf(
				"parsing optional sequence at offset %d but did not end with %q or %q",
				p.offset,
				']',
				"/)",
			)
		}
		p.offset += width
	}
	return definitionsList, nil
}

func (p *Parser) parseRepeatedSequence() (DefinitionsList, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	if char != '{' {
		if char != '(' {
			return nil, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not start with %q or %q",
				p.offset,
				'{',
				"(:",
			)
		}
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != ':' {
			return nil, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not start with %q or %q",
				p.offset,
				'{',
				"(:",
			)
		}
		p.offset += width
	}
	definitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	if char != '}' {
		if char != ':' {
			return nil, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not end with %q or %q",
				p.offset,
				'}',
				":)",
			)
		}
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != ')' {
			return nil, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not end with %q or %q",
				p.offset,
				'}',
				":)",
			)
		}
		p.offset += width
	}
	return definitionsList, nil
}

func (p *Parser) parseSpecialSequence() (string, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '?' {
		return "", fmt.Errorf(
			"parsing special sequence at offset %d but did not start with %q",
			p.offset,
			'?',
		)
	}
	p.offset += width
	startOffset := p.offset
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	for ; char != '?'; p.offset += width {
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	return p.source[startOffset : p.offset-width], nil
}

func (p *Parser) parseGroupedSequence() (DefinitionsList, error) {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	if char != '(' {
		return nil, fmt.Errorf(
			"parsing grouped sequence at offset %d but did not start with %q",
			p.offset,
			'(',
		)
	}
	definitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	if char != ')' {
		return nil, fmt.Errorf(
			"parsing grouped sequence at offset %d but did not end with %q",
			p.offset,
			')',
		)
	}
	p.offset += width
	return definitionsList, nil
}

func (p *Parser) parseTerminal() (string, error) {
	p.skipWhitespace()
	terminatingChar, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if terminatingChar != '\'' && terminatingChar != '"' {
		return "", fmt.Errorf(
			"parsing terminal at offset %d but did not start with %q or %q",
			p.offset,
			'\'',
			'"',
		)
	}
	p.offset += width
	startOffset := p.offset
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	for ; char != terminatingChar; p.offset += width {
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
	return p.source[startOffset : p.offset-width], nil
}

func (p *Parser) skipWhitespace() {
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for unicode.IsSpace(char) {
		p.offset += width
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
}
