package ebnf

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

func ParseSyntax(input string) (Syntax, error) {
	var syntax Syntax
	rule, offset, err := parseRule(input)
	if err != nil {
		return Syntax{}, err
	}
	syntax.Rules = append(syntax.Rules, rule)
	offset += skipWhitespace(input[offset:])
	for input[offset:] != "" {
		rule, width, err := parseRule(input[offset:])
		if err != nil {
			return Syntax{}, err
		}
		syntax.Rules = append(syntax.Rules, rule)
		offset += width
		offset += skipWhitespace(input[offset:])
	}
	return syntax, nil
}

func parseRule(input string) (Rule, int, error) {
	// Remove leading whitespace
	offset := skipWhitespace(input)
	// Look for start of meta identifier (letter)
	char, _ := utf8.DecodeRuneInString(input[offset:])
	if !unicode.IsLetter(char) {
		return Rule{}, 0, fmt.Errorf(
			"rule meta identifier does not start with letter, starts with %q",
			char,
		)
	}
	// Parse the meta identifier
	metaIdentifier, next, err := parseMetaIdentifier(input[offset:])
	if err != nil {
		return Rule{}, 0, err
	}
	offset += next
	rule := Rule{MetaIdentifier: metaIdentifier}
	// Remove leading whitespace
	offset += skipWhitespace(input[offset:])
	// Look for "=" character and remove it
	char, width := utf8.DecodeRuneInString(input[offset:])
	if char != '=' {
		return Rule{}, 0, fmt.Errorf(
			"next non-whitespace character after rule meta identifier should be %q but character at offset %d was not",
			'=',
			offset,
		)
	}
	offset += width
	// Parse a definitions list
	defintitionsList, width, err := parseDefinitionsList(input[offset:])
	if err != nil {
		return Rule{}, 0, err
	}
	offset += width
	rule.Definitions = defintitionsList
	// Look for terminating character
	offset += skipWhitespace(input[offset:])
	char, width = utf8.DecodeRuneInString(input[offset:])
	if char != ';' && char != '.' {
		return Rule{}, 0, fmt.Errorf(
			"no terminator symbol (one of %q or %q) found at end of syntax rule at offset %d",
			'.',
			';',
			offset,
		)
	}
	return rule, offset + width, nil
}

func skipWhitespace(input string) int {
	offset := 0
	char, width := utf8.DecodeRuneInString(input)
	for unicode.IsSpace(char) {
		offset += width
		char, width = utf8.DecodeRuneInString(input[offset:])
	}
	return offset
}

func parseMetaIdentifier(input string) (string, int, error) {
	offset := skipWhitespace(input)
	char, width := utf8.DecodeRuneInString(input)
	if !unicode.IsLetter(char) {
		return "", 0, fmt.Errorf(
			"parsing meta identifier at offset %d but first character was not letter",
			offset,
		)
	}
	startOffset := offset
	for unicode.IsLetter(char) || unicode.IsDigit(char) {
		offset += width
		char, width = utf8.DecodeRuneInString(input[offset:])
	}
	return input[startOffset:offset], offset, nil
}

func parseDefinitionsList(input string) (DefinitionsList, int, error) {
	var definitionsList DefinitionsList
	definition, offset, err := parseDefinition(input)
	if err != nil {
		return nil, 0, err
	}
	definitionsList = append(definitionsList, definition)
	offset += skipWhitespace(input[offset:])
	next, width := utf8.DecodeRuneInString(input[offset:])
	for next == '|' || next == '/' || next == '!' {
		if next == '/' {
			if next2, _ := utf8.DecodeRuneInString(input[offset+width:]); next2 == ')' {
				// "/)" is the end of an optional sequence
				break
			}
		}
		offset += width
		definition, width, err = parseDefinition(input[offset:])
		if err != nil {
			return nil, 0, err
		}
		definitionsList = append(definitionsList, definition)
		offset += width
		next, width = utf8.DecodeRuneInString(input[offset:])
	}
	return definitionsList, offset, nil
}

func parseDefinition(input string) (Definition, int, error) {
	var definition Definition
	term, offset, err := parseTerm(input)
	if err != nil {
		return Definition{}, 0, err
	}
	definition.Terms = append(definition.Terms, term)
	offset += skipWhitespace(input[offset:])
	next, width := utf8.DecodeRuneInString(input[offset:])
	for next == ',' {
		offset += width
		term, width, err = parseTerm(input[offset:])
		if err != nil {
			return Definition{}, 0, err
		}
		definition.Terms = append(definition.Terms, term)
		offset += width
		next, width = utf8.DecodeRuneInString(input[offset:])
	}
	return definition, offset, nil
}

func parseTerm(input string) (Term, int, error) {
	var term Term
	factor, offset, err := parseFactor(input)
	if err != nil {
		return Term{}, 0, err
	}
	term.Factor = factor
	offset += skipWhitespace(input[offset:])
	next, width := utf8.DecodeRuneInString(input[offset:])
	if next == '-' {
		offset += width
		offset += skipWhitespace(input[offset:])
		exception, width, err := parseFactor(input[offset:])
		if err != nil {
			return Term{}, 0, err
		}
		offset += width
		term.Exception = exception
	}
	return term, offset, nil
}

func parseFactor(input string) (Factor, int, error) {
	factor := Factor{Repetitions: -1}
	offset := skipWhitespace(input)
	char, _ := utf8.DecodeRuneInString(input[offset:])
	if unicode.IsDigit(char) {
		integer, width, err := parseInteger(input[offset:])
		if err != nil {
			return Factor{}, 0, err
		}
		offset += width
		factor.Repetitions = integer
		offset += skipWhitespace(input[offset:])
		char, width := utf8.DecodeRuneInString(input[offset:])
		if char != '*' {
			return Factor{}, 0, fmt.Errorf(
				"parsing factor had integer for repetitions but non-whitespace character after was not %q at offset %d",
				'*',
				offset,
			)
		}
		offset += width
	}
	primary, width, err := parsePrimary(input[offset:])
	if err != nil {
		return Factor{}, 0, err
	}
	offset += width
	factor.Primary = primary
	return factor, offset, nil
}

func parseInteger(input string) (int, int, error) {
	offset := skipWhitespace(input)
	input = input[offset:]
	startOffset := offset
	char, width := utf8.DecodeRuneInString(input[offset:])
	if !unicode.IsDigit(char) {
		return 0, 0, fmt.Errorf(
			"parsing integer but first character was not digit at offset %d",
			offset,
		)
	}
	offset += width
	for ; unicode.IsDigit(char); offset += width {
		char, width = utf8.DecodeRuneInString(input[offset:])
	}
	parsedInt, err := strconv.Atoi(input[startOffset : offset-width])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"parsed integer could not be converted to integer type at offset %d",
			startOffset,
		)
	}
	return parsedInt, offset, nil
}

func parsePrimary(input string) (Primary, int, error) {
	// Remove leading whitespace
	offset := skipWhitespace(input)
	input = input[offset:]
	primary := Primary{}
	var err error
	var width int
	// Parse one of:
	// optional sequence
	handleOptionalSequence := func() error {
		var optionalSequence DefinitionsList
		optionalSequence, width, err = parseOptionalSequence(input)
		if err != nil {
			return err
		}
		primary.OptionalSequence = optionalSequence
		return nil
	}
	// repeated sequence
	handleRepeatedSequence := func() error {
		var repeatedSequence DefinitionsList
		repeatedSequence, width, err = parseRepeatedSequence(input)
		if err != nil {
			return err
		}
		primary.RepeatedSequence = repeatedSequence
		return nil
	}
	// special sequence
	// grouped sequence
	// meta identifier
	// terminal
	// empty
	char, width := utf8.DecodeRuneInString(input)
	switch {
	case char == '[':
		if err := handleOptionalSequence(); err != nil {
			return Primary{}, 0, err
		}
	case char == '{':
		if err := handleRepeatedSequence(); err != nil {
			return Primary{}, 0, err
		}
	case char == '?':
		var specialSequence string
		specialSequence, width, err = parseSpecialSequence(input)
		if err != nil {
			return Primary{}, 0, err
		}
		primary.SpecialSequence = specialSequence
	case char == '(':
		next, _ := utf8.DecodeRuneInString(input[width:])
		switch next {
		case '/':
			if err := handleOptionalSequence(); err != nil {
				return Primary{}, 0, err
			}
		case ':':
			if err := handleRepeatedSequence(); err != nil {
				return Primary{}, 0, err
			}
		default:
			var groupedSequence DefinitionsList
			groupedSequence, width, err = parseGroupedSequence(input)
			if err != nil {
				return Primary{}, 0, err
			}
			primary.GroupedSequence = groupedSequence
		}
	case unicode.IsLetter(char):
		var metaIdentifier string
		metaIdentifier, width, err = parseMetaIdentifier(input)
		if err != nil {
			return Primary{}, 0, err
		}
		primary.MetaIdentifier = metaIdentifier
	case char == '\'':
		fallthrough
	case char == '"':
		var terminal string
		terminal, width, err = parseTerminal(input)
		if err != nil {
			return Primary{}, 0, err
		}
		primary.Terminal = terminal
	default:
		width = 0
		primary.Empty = true
	}
	return primary, offset + width, nil
}

func parseOptionalSequence(input string) (DefinitionsList, int, error) {
	offset := skipWhitespace(input)
	char, width := utf8.DecodeRuneInString(input[offset:])
	offset += width
	if char != '[' {
		if char != '(' {
			return nil, 0, fmt.Errorf(
				"parsing optional sequence at offset %d but did not start with %q or %q",
				offset,
				'[',
				"(/",
			)
		}
		char, width := utf8.DecodeRuneInString(input[offset:])
		if char != '/' {
			return nil, 0, fmt.Errorf(
				"parsing optional sequence at offset %d but did not start with %q or %q",
				offset,
				'[',
				"(/",
			)
		}
		offset += width
	}
	definitionsList, width, err := parseDefinitionsList(input[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += width
	offset += skipWhitespace(input[offset:])
	char, width = utf8.DecodeRuneInString(input[offset:])
	offset += width
	if char != ']' {
		if char != '/' {
			return nil, 0, fmt.Errorf(
				"parsing optional sequence at offset %d but did not end with %q or %q",
				offset,
				']',
				"/)",
			)
		}
		char, width := utf8.DecodeRuneInString(input[offset:])
		if char != ')' {
			return nil, 0, fmt.Errorf(
				"parsing optional sequence at offset %d but did not end with %q or %q",
				offset,
				']',
				"/)",
			)
		}
		offset += width
	}
	return definitionsList, offset, nil
}

func parseRepeatedSequence(input string) (DefinitionsList, int, error) {
	offset := skipWhitespace(input)
	char, width := utf8.DecodeRuneInString(input[offset:])
	offset += width
	if char != '{' {
		if char != '(' {
			return nil, 0, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not start with %q or %q",
				offset,
				'{',
				"(:",
			)
		}
		char, width := utf8.DecodeRuneInString(input[offset:])
		if char != ':' {
			return nil, 0, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not start with %q or %q",
				offset,
				'{',
				"(:",
			)
		}
		offset += width
	}
	definitionsList, width, err := parseDefinitionsList(input[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += width
	offset += skipWhitespace(input[offset:])
	char, width = utf8.DecodeRuneInString(input[offset:])
	offset += width
	if char != '}' {
		if char != ':' {
			return nil, 0, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not end with %q or %q",
				offset,
				'}',
				":)",
			)
		}
		char, width := utf8.DecodeRuneInString(input[offset:])
		if char != ')' {
			return nil, 0, fmt.Errorf(
				"parsing repeated sequence at offset %d but did not end with %q or %q",
				offset,
				'}',
				":)",
			)
		}
		offset += width
	}
	return definitionsList, offset, nil
}

func parseSpecialSequence(input string) (string, int, error) {
	offset := skipWhitespace(input)
	input = input[offset:]
	char, width := utf8.DecodeRuneInString(input)
	if char != '?' {
		return "", 0, fmt.Errorf(
			"parsing special sequence at offset %d but did not start with %q",
			offset,
			'?',
		)
	}
	offset += width
	startOffset := offset
	char, width = utf8.DecodeRuneInString(input[offset:])
	offset += width
	for ; char != '?'; offset += width {
		char, width = utf8.DecodeRuneInString(input[offset:])
	}
	return input[startOffset : offset-width], offset, nil
}

func parseGroupedSequence(input string) (DefinitionsList, int, error) {
	offset := skipWhitespace(input)
	char, width := utf8.DecodeRuneInString(input[offset:])
	offset += width
	if char != '(' {
		return nil, 0, fmt.Errorf(
			"parsing grouped sequence at offset %d but did not start with %q",
			offset,
			'(',
		)
	}
	definitionsList, width, err := parseDefinitionsList(input[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += width
	offset += skipWhitespace(input[offset:])
	char, width = utf8.DecodeRuneInString(input[offset:])
	if char != ')' {
		return nil, 0, fmt.Errorf(
			"parsing grouped sequence at offset %d but did not end with %q",
			offset,
			')',
		)
	}
	offset += width
	return definitionsList, offset, nil
}

func parseTerminal(input string) (string, int, error) {
	offset := skipWhitespace(input)
	input = input[offset:]
	terminatingChar, width := utf8.DecodeRuneInString(input)
	if terminatingChar != '\'' && terminatingChar != '"' {
		return "", 0, fmt.Errorf(
			"parsing terminal at offset %d but did not start with %q or %q",
			offset,
			'\'',
			'"',
		)
	}
	offset += width
	startOffset := offset
	char, width := utf8.DecodeRuneInString(input[offset:])
	offset += width
	for ; char != terminatingChar; offset += width {
		char, width = utf8.DecodeRuneInString(input[offset:])
	}
	return input[startOffset : offset-width], offset, nil
}
