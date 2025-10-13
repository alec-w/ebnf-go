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
	// Parse one of:
	// optional sequence
	// repeated sequence
	// special sequence
	// grouped sequence
	// meta identifier
	// terminal
	// empty
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	switch {
	case char == '[':
		optionalSequence, err := p.parseOptionalSequence()
		if err != nil {
			return Primary{}, err
		}
		primary.OptionalSequence = optionalSequence
	case char == '{':
		repeatedSequence, err := p.parseRepeatedSequence()
		if err != nil {
			return Primary{}, err
		}
		primary.RepeatedSequence = repeatedSequence
	case char == '?':
		specialSequence, err := p.parseSpecialSequence()
		if err != nil {
			return Primary{}, err
		}
		primary.SpecialSequence = specialSequence
	case char == '(':
		next, _ := utf8.DecodeRuneInString(p.source[p.offset+width:])
		switch next {
		case '/':
			optionalSequence, err := p.parseOptionalSequence()
			if err != nil {
				return Primary{}, err
			}
			primary.OptionalSequence = optionalSequence
		case ':':
			repeatedSequence, err := p.parseRepeatedSequence()
			if err != nil {
				return Primary{}, err
			}
			primary.RepeatedSequence = repeatedSequence
		default:
			groupedSequence, err := p.parseGroupedSequence()
			if err != nil {
				return Primary{}, err
			}
			primary.GroupedSequence = groupedSequence
		}
	case unicode.IsLetter(char):
		metaIdentifier, err := p.parseMetaIdentifier()
		if err != nil {
			return Primary{}, err
		}
		primary.MetaIdentifier = metaIdentifier
	case char == '\'':
		fallthrough
	case char == '"':
		terminal, err := p.parseTerminal()
		if err != nil {
			return Primary{}, err
		}
		primary.Terminal = terminal
	default:
		primary.Empty = true
	}
	return primary, nil
}

func (p *Parser) parseOptionalSequence() (DefinitionsList, error) {
	return p.parseWrappedDefinitionsList(
		"optional sequence",
		[][]rune{{'['}, {'(', '/'}},
		[][]rune{{']'}, {'/', ')'}},
	)
}

func (p *Parser) parseRepeatedSequence() (DefinitionsList, error) {
	return p.parseWrappedDefinitionsList(
		"repeated sequence",
		[][]rune{{'{'}, {'(', ':'}},
		[][]rune{{'}'}, {':', ')'}},
	)
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
	return p.parseWrappedDefinitionsList(
		"repeated sequence",
		[][]rune{{'('}},
		[][]rune{{')'}},
	)
}

func (p *Parser) parseWrappedDefinitionsList(
	sequenceName string,
	startIdentifiers, endIdentifiers [][]rune,
) (DefinitionsList, error) {
	parseEnclosingCharacters := func(position string, identifiers [][]rune) error {
		p.skipWhitespace()
		var found bool
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		for _, identifier := range identifiers {
			totalWidth := 0
			for _, identifierChar := range identifier {
				if identifierChar != char {
					break
				}
				totalWidth += width
				char, width = utf8.DecodeRuneInString(p.source[p.offset+totalWidth:])
			}
			if totalWidth == len(identifier) {
				p.offset += totalWidth
				found = true
				break
			}
		}
		if found {
			return nil
		}
		var identifierStrings []string
		for _, identifier := range identifiers {
			identifierStrings = append(identifierStrings, string(identifier))
		}
		var errSuffix string
		if len(identifierStrings) == 0 {
			errSuffix = "but no " + position + " identifiers were supplied"
		} else {
			errSuffix = fmt.Sprintf("%q", identifierStrings[len(identifierStrings)-1])
			identifierStrings = identifierStrings[:len(identifierStrings)-1]
			if len(identifierStrings) > 0 {
				errSuffix = fmt.Sprintf("%q or "+errSuffix, identifierStrings[len(identifierStrings)-1])
			}
			identifierStrings = identifierStrings[:len(identifierStrings)-1]
			if len(identifierStrings) > 0 {
				for i := len(identifierStrings) - 1; i >= 0; i-- {
					errSuffix = fmt.Sprintf("%q, "+errSuffix, identifierStrings[i])
				}
			}
			errSuffix = "but did not " + position + " with " + errSuffix
		}
		return fmt.Errorf(
			"parsing %s at offset %d %s",
			sequenceName,
			p.offset,
			errSuffix,
		)
	}
	if err := parseEnclosingCharacters("start", startIdentifiers); err != nil {
		return nil, err
	}
	definitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return nil, err
	}
	if err := parseEnclosingCharacters("end", endIdentifiers); err != nil {
		return nil, err
	}
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
