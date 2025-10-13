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

// Parse is the entrypoint of the parser.
//
// Given a source EBNF grammar it produces a structured representation of it.
func (p *Parser) Parse(source string) (Syntax, error) {
	p.source = source
	p.offset = 0
	return p.parseSyntax()
}

func (p *Parser) parseSyntax() (Syntax, error) {
	// A syntax is made up of one or more rules.
	// So parse one rule...
	var syntax Syntax
	rule, err := p.parseRule()
	if err != nil {
		return Syntax{}, err
	}
	syntax.Rules = append(syntax.Rules, rule)
	p.skipWhitespace()
	// ...then optionally parse more if the entire grammar has not been parsed.
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
	// A rule is made up of a meta identifier followed by a literal "=" then a list of definitions, then a terminating
	// symbol (";" or ".")
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
	metaIdentifier := p.parseMetaIdentifier()
	rule := Rule{MetaIdentifier: metaIdentifier}
	// Remove leading whitespace
	p.skipWhitespace()
	// Look for "=" character
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

func (p *Parser) parseMetaIdentifier() string {
	p.skipWhitespace()
	// A meta identifier is a sequence of letters and digits starting with a letter.
	// This assumes that first character has already checked to be a letter.
	// Since this is internal to the parser this should be checked there to avoid unreachable error handling code here
	startOffset := p.offset
	for {
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			break
		}
		p.offset += width
	}
	return p.source[startOffset:p.offset]
}

func (p *Parser) parseDefinitionsList() (DefinitionsList, error) {
	// A defintions list is a sequence of one or more definitions separated by "|", "/" or "!".
	// So parse the first definition...
	definition, err := p.parseDefinition()
	if err != nil {
		return nil, err
	}
	definitionsList := DefinitionsList{definition}
	p.skipWhitespace()
	// ...then optionally parse additional definitions.
	next, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for next == '|' || next == '/' || next == '!' {
		if next == '/' {
			if next2, _ := utf8.DecodeRuneInString(p.source[p.offset+width:]); next2 == ')' {
				// "/)" is the end of an optional sequence, which contains a definition list
				// so have to peek twice here to check that and avoid swallowing it
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
	// A definition is a sequence of one or more terms separated by ","
	// So parse one term...
	term, err := p.parseTerm()
	if err != nil {
		return Definition{}, err
	}
	definition := Definition{Terms: []Term{term}}
	p.skipWhitespace()
	// ...then optionally parse additional terms
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
	// A term is a factor, then with an optional exception (also a factor) preceded by a literal "-"
	// So parse a factor...
	factor, err := p.parseFactor()
	if err != nil {
		return Term{}, err
	}
	term := Term{Factor: factor}
	term.Factor = factor
	// ...then optionally parse an exception
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
	// A factor is a primary preceded by an optional integer number of repetitions (followed by a literal "*")
	// By the spec, a repetitions of 0 is allowed (although pointless), so default (unspecified) to -1 to distinguish
	// that case.
	factor := Factor{Repetitions: -1}
	// So optionally parse a repetition count...
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
	// ...then parse a primary
	primary, err := p.parsePrimary()
	if err != nil {
		return Factor{}, err
	}
	factor.Primary = primary
	return factor, nil
}

func (p *Parser) parseInteger() (int, error) {
	// An integer is a sequence of one of more digits.
	// This assumes the first character is a digit as this is internal to the parser so this should have
	// already been checked to avoid unrreachable error handling code here.
	p.skipWhitespace()
	startOffset := p.offset
	for {
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if !unicode.IsDigit(char) {
			break
		}
		p.offset += width
	}
	parsedInt, err := strconv.Atoi(p.source[startOffset:p.offset])
	// The spec allows unsized integers, for simplicity this only allows up to 2^63-1, which should be enough for all
	// practical grammars.
	// The simplest example of a grammar that used a value larger than this (and doesn't arbitrarily define the empty
	// string) would be
	// root = 9223372036854775808 * "0" ;
	// which (if encoding "0" in a single byte) would require exabytes of text to have required the 2^63 repetitions.
	if err != nil {
		return 0, fmt.Errorf(
			"parsed integer could not be converted to integer type at offset %d",
			startOffset,
		)
	}
	return parsedInt, nil
}

func (p *Parser) parsePrimary() (Primary, error) {
	// A primary is one of an optional sequence, a repeated sequence, a special sequence, a grouped sequence, a meta
	// identifier, a terminal, or empty.
	// To determine which one should be matched the next character is inspected
	p.skipWhitespace()
	primary := Primary{}
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
		// "(" can denote the start of an optional sequence, repeated sequence or grouped sequence depending on the
		// next character
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
		primary.MetaIdentifier = p.parseMetaIdentifier()
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
	// An optional sequence is a definitions list wrapped in either [...] or (/.../)
	// The spec says that grammars should only use one of these throughout, but the parser is more lenient
	// as it is possible to support that without changing the parsing behaviour for grammars that use one set of symbols.
	return p.parseWrappedDefinitionsList(
		"optional sequence",
		[][]rune{{'['}, {'(', '/'}},
		[][]rune{{']'}, {'/', ')'}},
	)
}

func (p *Parser) parseRepeatedSequence() (DefinitionsList, error) {
	// A repeated sequence is a definitions list wrapped in either {...} or (:...:)
	// The spec says that grammars should only use one of these throughout, but the parser is more lenient
	// as it is possible to support that without changing the parsing behaviour for grammars that use one set of symbols.
	return p.parseWrappedDefinitionsList(
		"repeated sequence",
		[][]rune{{'{'}, {'(', ':'}},
		[][]rune{{'}'}, {':', ')'}},
	)
}

func (p *Parser) parseSpecialSequence() (string, error) {
	// A special sequence is any sequence of characters apart from "?" wrapped in ?...?.
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
	// A grouped sequence is a definitions list wrapped in parentheses (...)
	return p.parseWrappedDefinitionsList(
		"repeated sequence",
		[][]rune{{'('}},
		[][]rune{{')'}},
	)
}

// parseWrappedDefinitionsList is a utility function used to parse repeated, optional and grouped sequences as the
// logic is the same for each because they are just definitions lists wrapped in different enclosing characters.
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
	// A terminal is any set of characters apart from single quotes, wrapped in single quotes,
	// or any set of characters apart from double quotes wrapped in double quotes.
	// Essentially '...' or "..." where the character used as the terminator does not appear inside.
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

// skipWhitespace is a utility function used to skip whitespace.
//
// The spec allows whitespace anywhere between the different components, whitespace is used to make a grammar easier
// to read but does not change its meaning.
func (p *Parser) skipWhitespace() {
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for unicode.IsSpace(char) {
		p.offset += width
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
}
