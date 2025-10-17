package ebnf

import (
	"strconv"
	"unicode"
	"unicode/utf8"
)

type Parser struct {
	source string
	offset int
	line   int
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
	p.line = 1

	return p.parseSyntax()
}

func (p *Parser) parseSyntax() (Syntax, error) {
	// A syntax is made up of one or more rules.
	// So parse one rule...
	var syntax Syntax
	// It is not possible to distinguish between comments on the syntax as a whole and comments on the first rule of
	// the syntax therefore comments at the start of the syntax will be attached to the first rule, so they are not
	// parsed here.
	var comments []string
	for p.isCommentStart() {
		comments = append(comments, p.parseComment())
	}
	rule, err := p.parseRule()
	if err != nil {
		return Syntax{}, err
	}
	rule.Comments = comments
	syntax.Rules = append(syntax.Rules, rule)
	p.skipWhitespace()
	// ...then optionally parse more if the entire grammar has not been parsed.
	for p.source[p.offset:] != "" {
		// We may have trailing comments or another rule (optionally preceded by comments)
		// So parse any comments...
		var comments []string
		for p.isCommentStart() {
			comments = append(comments, p.parseComment())
		}
		p.skipWhitespace()
		// ...then see if we've reached the end (so these were trailing comments)...
		if p.source[p.offset:] == "" {
			syntax.TrailingComments = comments

			break
		}
		// ...otherwise those were the comments preceding the next rule
		rule, err := p.parseRule()
		if err != nil {
			return Syntax{}, err
		}
		rule.Comments = comments
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
		return Rule{}, &ParseError{
			Msg:    "expected rule meta identifier start character (letter)",
			Offset: p.offset,
		}
	}
	rule := Rule{Line: p.line}
	// Parse the meta identifier
	rule.MetaIdentifier = p.parseMetaIdentifier()
	for p.isCommentStart() {
		rule.Comments = append(rule.Comments, p.parseComment())
	}
	// Remove leading whitespace
	p.skipWhitespace()
	// Look for "=" character
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char != '=' {
		return Rule{}, &ParseError{Msg: "expected defining symbol ('=')", Offset: p.offset}
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
		return Rule{}, &ParseError{Msg: "expected terminator symbol ('.' or ';')", Offset: p.offset}
	}
	p.offset += width

	return rule, nil
}

func (p *Parser) parseMetaIdentifier() string {
	p.skipWhitespace()
	// A meta identifier is a sequence of letters and digits starting with a letter.
	// Preceding/Tailing whitespace is allowed as is whitespace between characters - this is all ignored.
	// This assumes that first character has already checked to be a letter.
	// Since this is internal to the parser this should be checked there to avoid unreachable error handling code here
	var metaIdentifier []rune
	for {
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			break
		}
		p.offset += width
		metaIdentifier = append(metaIdentifier, char)
		p.skipWhitespace()
	}

	return string(metaIdentifier)
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
	// Optionally parse any preceding comments
	for p.isCommentStart() {
		factor.Comments = append(factor.Comments, p.parseComment())
	}
	// So optionally parse a repetition count...
	p.skipWhitespace()
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	if unicode.IsDigit(char) {
		integer, err := p.parseInteger()
		if err != nil {
			return Factor{}, err
		}
		factor.Repetitions = integer
		// Optionally parse any comments after the number of repetitions
		for p.isCommentStart() {
			factor.Comments = append(factor.Comments, p.parseComment())
		}
		p.skipWhitespace()
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		if char != '*' {
			return Factor{}, &ParseError{Msg: "expected repetition symbol ('*')", Offset: p.offset}
		}
		p.offset += width
		// Optionally parse any comments after the repetitions
		for p.isCommentStart() {
			factor.Comments = append(factor.Comments, p.parseComment())
		}
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
	// The simplest example of a grammar that used a value larger than this would be
	// root = 9223372036854775808 * "0" ;
	// which (if encoding "0" in a single byte) would require exabytes of text to have required the 2^63 repetitions.
	if err != nil {
		return 0, &ParseError{
			Msg:     "integer could not be parsed (max integer size is 2^63-1)",
			Offset:  p.offset,
			wrapped: err,
		}
	}

	return parsedInt, nil
}

func (p *Parser) parsePrimary() (Primary, error) {
	// A primary is one of an optional sequence, a repeated sequence, a special sequence, a grouped sequence, a meta
	// identifier, a terminal, or empty.
	// To determine which one should be matched the next character is inspected
	p.skipWhitespace()
	primary := Primary{}
	var err error
	char, _ := utf8.DecodeRuneInString(p.source[p.offset:])
	switch {
	case char == '[':
		var optionalSequence DefinitionsList
		optionalSequence, err = p.parseOptionalSequence()
		primary.OptionalSequence = optionalSequence
	case char == '{':
		var repeatedSequence DefinitionsList
		repeatedSequence, err = p.parseRepeatedSequence()
		primary.RepeatedSequence = repeatedSequence
	case char == '?':
		primary.SpecialSequence = p.parseSpecialSequence()
	case char == '(':
		primary, err = p.parseParenthisedSequence()
	case unicode.IsLetter(char):
		primary.MetaIdentifier = p.parseMetaIdentifier()
	case char == '\'':
		fallthrough
	case char == '"':
		primary.Terminal = p.parseTerminal()
	default:
		primary.Empty = true
	}

	return primary, err
}

func (p *Parser) parseParenthisedSequence() (Primary, error) {
	// This assumes the character at offset is "(", which should have already been checked by the caller as this is
	// internal to the parser.
	var primary Primary
	_, width := utf8.DecodeRuneInString(p.source[p.offset:])
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

	return primary, nil
}

func (p *Parser) parseOptionalSequence() (DefinitionsList, error) {
	// An optional sequence is a definitions list wrapped in either [...] or (/.../)
	// The spec says that grammars should only use one of these throughout, but the parser is more lenient
	// as it is possible to support that without changing the parsing behaviour for grammars that use one set of symbols.
	return p.parseWrappedDefinitionsList(
		[][]rune{{'['}, {'(', '/'}},
		[][]rune{{']'}, {'/', ')'}},
	)
}

func (p *Parser) parseRepeatedSequence() (DefinitionsList, error) {
	// A repeated sequence is a definitions list wrapped in either {...} or (:...:)
	// The spec says that grammars should only use one of these throughout, but the parser is more lenient
	// as it is possible to support that without changing the parsing behaviour for grammars that use one set of symbols.
	return p.parseWrappedDefinitionsList(
		[][]rune{{'{'}, {'(', ':'}},
		[][]rune{{'}'}, {':', ')'}},
	)
}

func (p *Parser) parseSpecialSequence() string {
	// A special sequence is any sequence of characters apart from "?" wrapped in ?...?.
	p.skipWhitespace()
	// This assumes the first non-whitespace character is "?" which should have been checked before calling this
	// function. As this is internal to the parser this allows the removal of unreachable error handling code.
	_, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	// Leading whitespace is ignored
	p.skipWhitespace()
	startOffset := p.offset
	for {
		var char rune
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
		p.offset += width
		if char == '?' {
			break
		}
		if char == '\n' {
			p.line++
		}
	}
	// Trailing whitespace is ignored
	endOffset := p.offset - width
	for {
		char, width := utf8.DecodeLastRuneInString(p.source[startOffset:endOffset])
		if !unicode.IsSpace(char) {
			break
		}
		endOffset -= width
	}

	return p.source[startOffset:endOffset]
}

func (p *Parser) parseGroupedSequence() (DefinitionsList, error) {
	// A grouped sequence is a definitions list wrapped in parentheses (...)
	return p.parseWrappedDefinitionsList(
		[][]rune{{'('}},
		[][]rune{{')'}},
	)
}

// parseWrappedDefinitionsList is a utility function used to parse repeated, optional and grouped sequences as the
// logic is the same for each because they are just definitions lists wrapped in different enclosing characters.
func (p *Parser) parseWrappedDefinitionsList(
	startIdentifiers, endIdentifiers [][]rune,
) (DefinitionsList, error) {
	// This assumes that the source at the current offset already starts with the one of the given start identifier#
	// sequences (after whitespace is ignored) since this is internal to the parser this avoids unreachable error
	// handling code.
	parseEnclosingCharacters := func(identifiers [][]rune) {
		p.skipWhitespace()
		char, width := utf8.DecodeRuneInString(p.source[p.offset:])
		for _, identifier := range identifiers {
			chars := 0
			totalWidth := 0
			for _, identifierChar := range identifier {
				if identifierChar != char {
					break
				}
				chars++
				totalWidth += width
				char, width = utf8.DecodeRuneInString(p.source[p.offset+totalWidth:])
			}
			if chars == len(identifier) {
				p.offset += totalWidth

				break
			}
		}
	}
	parseEnclosingCharacters(startIdentifiers)
	definitionsList, err := p.parseDefinitionsList()
	if err != nil {
		return nil, err
	}
	parseEnclosingCharacters(endIdentifiers)

	return definitionsList, nil
}

func (p *Parser) parseTerminal() string {
	// A terminal is any set of characters apart from single quotes, wrapped in single quotes,
	// or any set of characters apart from double quotes wrapped in double quotes.
	// Essentially '...' or "..." where the character used as the terminator does not appear inside.
	p.skipWhitespace()
	// This assumes the next character is either single quote or double quote which should have been checked already
	// as this is internal to the parser this avoids unreachable error handling code.
	terminatingChar, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	startOffset := p.offset
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '\n' {
		p.line++
	}
	p.offset += width
	for ; char != terminatingChar; p.offset += width {
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
		if char == '\n' {
			p.line++
		}
	}

	return p.source[startOffset : p.offset-width]
}

func (p *Parser) parseComment() string {
	// A comment is a repeated sequence of comment symbols wrapped in parentheses and stars (*...*).
	p.skipWhitespace()
	// This assumes the first non-whitespace characters are "(*" which should have been checked before calling this
	// function. As this is internal to the parser this allows the removal of unreachable error handling code.
	_, width := utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	_, width = utf8.DecodeRuneInString(p.source[p.offset:])
	p.offset += width
	// Leading whitespace is ignored
	p.skipWhitespace()
	startOffset := p.offset
	for {
		p.skipWhitespace()
		if char, width := utf8.DecodeRuneInString(p.source[p.offset:]); char == '*' {
			next, nextWidth := utf8.DecodeRuneInString(p.source[p.offset+width:])
			if next == ')' {
				p.offset += width + nextWidth
				// Trailing whitespace is ignored
				endOffset := p.offset - (width + nextWidth)
				for {
					char, width := utf8.DecodeLastRuneInString(p.source[startOffset:endOffset])
					if !unicode.IsSpace(char) {
						break
					}
					endOffset -= width
				}

				return p.source[startOffset:endOffset]
			}
		}
		p.parseCommentSymbol()
	}
}

func (p *Parser) parseCommentSymbol() {
	// A comment symbol is a comment, a terminal, a special sequence or any other character.
	// This means that comments can enclose other comments, but the inner comments must be correctly terminated
	// and comments can contain quoted strings, but they must be correctly terminated, and comments can include
	// special sequences, but they must be correctly terminated.
	// This function doesn't return anything, but simply advances the offset forwards, so the outermost comment is
	// just stored as a sequence of characters on the final parsed syntax.
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	switch char {
	case '(':
		next, _ := utf8.DecodeRuneInString(p.source[p.offset+width:])
		if next == '*' {
			p.parseComment()
		} else {
			p.offset += width
		}
	case '\'':
		fallthrough
	case '"':
		p.parseTerminal()
	case '?':
		p.parseSpecialSequence()
	default:
		if char == '\n' {
			p.line++
		}
		p.offset += width
	}
}

// isCommentStart is a utility function used to check if the next non whitespace character is a comment start symbol.
func (p *Parser) isCommentStart() bool {
	p.skipWhitespace()
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	if char == '(' {
		next, _ := utf8.DecodeRuneInString(p.source[p.offset+width:])
		if next == '*' {
			return true
		}
	}

	return false
}

// skipWhitespace is a utility function used to skip whitespace.
//
// The spec allows whitespace anywhere between the different components, whitespace is used to make a grammar easier
// to read but does not change its meaning.
func (p *Parser) skipWhitespace() {
	char, width := utf8.DecodeRuneInString(p.source[p.offset:])
	for unicode.IsSpace(char) {
		if char == '\n' {
			p.line++
		}
		p.offset += width
		char, width = utf8.DecodeRuneInString(p.source[p.offset:])
	}
}
