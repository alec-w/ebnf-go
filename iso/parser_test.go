package iso_test

import (
	"strings"
	"testing"

	"github.com/alec-w/ebnf-go/iso"
)

func assertSlicesEqual[T any](
	t *testing.T,
	expected, actual []T,
	itemPlural, itemSingular string,
	itemComparison func(*testing.T, T, T) bool,
) bool {
	t.Helper()
	var failed bool
	if len(expected) != len(actual) {
		t.Logf("Expected %d %s. Got %d.", len(expected), itemPlural, len(actual))
		t.Fail()

		return false
	}
	for i, expectedItem := range expected {
		if !itemComparison(t, expectedItem, actual[i]) {
			t.Logf(
				"%s %d did not match expected",
				strings.ToTitle(itemSingular[0:1])+itemSingular[1:],
				i+1,
			)
			t.Fail()
			failed = true
		}
	}

	return !failed
}

func assertDefinitionsListsEqual(t *testing.T, expected, actual iso.DefinitionsList) bool {
	t.Helper()

	return assertSlicesEqual(
		t,
		expected,
		actual,
		"definitions",
		"definition",
		assertDefinitionsEqual,
	)
}

func assertSyntaxesEqual(t *testing.T, expected, actual iso.Syntax) bool {
	t.Helper()
	if assertSlicesEqual(t, expected.Rules, actual.Rules, "rules", "rule", assertRulesEqual) {
		return true
	}
	t.Log("Syntax rules were not equal")
	t.Fail()

	return false
}

func assertCommentsEqual(t *testing.T, expected, actual string) bool {
	t.Helper()
	if expected == actual {
		return true
	}
	t.Logf("Expected comment %q. Got %q.", expected, actual)
	t.Fail()

	return false
}

func assertRulesEqual(t *testing.T, expected, actual iso.Rule) bool {
	t.Helper()
	var failed bool
	if expected.MetaIdentifier != actual.MetaIdentifier {
		t.Logf(
			"Expected rule meta identifier %q. Got %q.",
			expected.MetaIdentifier,
			actual.MetaIdentifier,
		)
		t.Fail()
		failed = true
	}
	ruleName := "Rule"
	if expected.MetaIdentifier != "" && expected.MetaIdentifier == actual.MetaIdentifier {
		ruleName += " \"" + expected.MetaIdentifier + "\""
	}
	if !assertSlicesEqual(
		t,
		expected.Comments,
		actual.Comments,
		"comments",
		"comment",
		assertCommentsEqual,
	) {
		t.Logf("%s comments were not equal", ruleName)
		t.Fail()
		failed = true
	}
	if expected.Line != actual.Line {
		t.Logf(
			"Expected %s to be on line %d. Got %d.",
			strings.ToLower(ruleName[0:1])+ruleName[1:],
			expected.Line,
			actual.Line,
		)
		t.Fail()
		failed = true
	}
	if assertDefinitionsListsEqual(t, expected.Definitions, actual.Definitions) {
		return !failed
	}
	t.Logf("%s definitions were not equal", ruleName)
	t.Fail()

	return false
}

func assertDefinitionsEqual(t *testing.T, expected, actual iso.Definition) bool {
	t.Helper()
	if assertSlicesEqual(t, expected.Terms, actual.Terms, "terms", "term", assertTermsEqual) {
		return true
	}
	t.Log("Definition terms were not equal")
	t.Fail()

	return false
}

func assertTermsEqual(t *testing.T, expected, actual iso.Term) bool {
	t.Helper()
	var failed bool
	if !assertFactorsEqual(t, expected.Factor, actual.Factor) {
		t.Log("Term factors were not equal")
		t.Fail()
		failed = true
	}
	if assertFactorsEqual(t, expected.Exception, actual.Exception) {
		return !failed
	}
	t.Log("Term exceptions were not equal")
	t.Fail()

	return false
}

func assertFactorsEqual(t *testing.T, expected, actual iso.Factor) bool {
	t.Helper()
	var failed bool
	if expected.Repetitions != actual.Repetitions {
		t.Logf(
			"Expected factor to have %d repetitions. Got %d.",
			expected.Repetitions,
			actual.Repetitions,
		)
		t.Fail()
		failed = true
	}
	if !assertSlicesEqual(
		t,
		expected.Comments,
		actual.Comments,
		"comments",
		"comment",
		assertCommentsEqual,
	) {
		t.Log("Factor comments were not equal")
		t.Fail()
		failed = true
	}
	if !assertPrimariesEqual(t, expected.Primary, actual.Primary) {
		t.Log("Factor primaries were not equal")
		t.Fail()
		failed = true
	}

	return !failed
}

func assertPrimariesEqual(t *testing.T, expected, actual iso.Primary) bool {
	t.Helper()
	var failed bool
	if !assertDefinitionsListsEqual(t, expected.OptionalSequence, actual.OptionalSequence) {
		t.Log("Primary optional sequences not equal")
		failed = true
	}
	if !assertDefinitionsListsEqual(t, expected.RepeatedSequence, actual.RepeatedSequence) {
		t.Log("Primary repeated sequences not equal")
		failed = true
	}
	if expected.SpecialSequence != actual.SpecialSequence {
		t.Logf(
			"Expected primary special sequence %q. Got %q.",
			expected.SpecialSequence,
			actual.SpecialSequence,
		)
		failed = true
	}
	if !assertDefinitionsListsEqual(t, expected.GroupedSequence, actual.GroupedSequence) {
		t.Log("Primary grouped sequences not equal")
		failed = true
	}
	if expected.MetaIdentifier != actual.MetaIdentifier {
		t.Logf(
			"Expected primary meta identifier %q. Got %q.",
			expected.MetaIdentifier,
			actual.MetaIdentifier,
		)
		failed = true
	}
	if expected.Terminal != actual.Terminal {
		t.Logf("Expected primary terminal %q. Got %q.", expected.Terminal, actual.Terminal)
		failed = true
	}
	if expected.Empty != actual.Empty {
		t.Logf("Expected primary empty %t. Got %t.", expected.Empty, actual.Empty)
		failed = true
	}
	if failed {
		t.Fail()
	}

	return !failed
}

//nolint:godox // todo is tracking work to be done
// TODO restructure test

//nolint:funlen // test to be restructured
func TestParseSyntax(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name           string
		grammar        string
		expectedSyntax iso.Syntax
	}{
		{
			name: "Positive integer definition",
			grammar: `
nonZeroDigit = "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
digit = "0" | nonZeroDigit ;
integer = "0" | nonZeroDigit, { digit } ;
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "nonZeroDigit",
						Line:           2,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "1"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "2"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "3"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "4"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "5"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "6"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "7"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "8"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "9"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "digit",
						Line:           3,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "0"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "nonZeroDigit",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "integer",
						Line:           4,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "0"},
										},
									},
								},
							},
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "nonZeroDigit"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "digit",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
				},
			},
		},
		{
			name: "Extended EBNF defined informally",
			grammar: `
SYNTAX = SYNTAX RULE, (: SYNTAX RULE :).
SYNTAX RULE
= META IDENTIFIER, '=', DEFINITIONS LIST, '.'.
DEFINITIONS LIST
= SINGLE DEFINITION,
(: '/', SINGLE DEFINITION :).
SINGLE DEFINITION = TERM, (: ',', TERM :).
TERM = FACTOR, (/ '-', EXCEPTION /).
EXCEPTION = FACTOR.
FACTOR = (/ INTEGER, '*' /), PRIMARY.
PRIMARY
= OPTIONAL SEQUENCE / REPEATED SEQUENCE
/ SPECIAL SEQUENCE / GROUPED SEQUENCE
/ META IDENTIFIER / TERMINAL / EMPTY.
EMPTY = .
OPTIONAL SEQUENCE = '(/', DEFINITIONS LIST, '/)'.
REPEATED SEQUENCE = '(:', DEFINITIONS LIST, ':)'.
GROUPED SEQUENCE = '(', DEFINITIONS LIST, ')'.
TERMINAL
= "'" , CHARACTER - "'",
(: CHARACTER - "'" :), "'"
/ '"' , CHARACTER - '"',
(: CHARACTER - '"' :), '"'.
META IDENTIFIER = LETTER, (: LETTER / DIGIT :).
INTEGER = DIGIT, (: DIGIT :).
SPECIAL SEQUENCE = '?', (: CHARACTER - '?' :), '?'.
COMMENT = '(*', (: COMMENT SYMBOL :), '*)'.
COMMENT SYMBOL
= COMMENT / TERMINAL / SPECIAL SEQUENCE
/ CHARACTER.`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "SYNTAX",
						Line:           2,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "SYNTAXRULE",
											},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												RepeatedSequence: iso.DefinitionsList{
													{
														Terms: []iso.Term{
															{
																Factor: iso.Factor{
																	Repetitions: -1,
																	Primary: iso.Primary{
																		MetaIdentifier: "SYNTAXRULE",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "SYNTAXRULE",
						Line:           3,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "METAIDENTIFIER",
											},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "="},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "DEFINITIONSLIST",
											},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "."},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "DEFINITIONSLIST",
						Line:           5,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "SINGLEDEFINITION",
											},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												RepeatedSequence: iso.DefinitionsList{
													{Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "/",
																},
															},
														},
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "SINGLEDEFINITION",
																},
															},
														},
													}},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "SINGLEDEFINITION",
						Line:           8,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "TERM"},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												RepeatedSequence: iso.DefinitionsList{
													{Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: ",",
																},
															},
														},
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "TERM",
																},
															},
														},
													}},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "TERM",
						Line:           9,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "FACTOR"},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												OptionalSequence: iso.DefinitionsList{
													{Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "-",
																},
															},
														},
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "EXCEPTION",
																},
															},
														},
													}},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "EXCEPTION",
						Line:           10,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "FACTOR"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "FACTOR",
						Line:           11,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												OptionalSequence: iso.DefinitionsList{
													{Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "INTEGER",
																},
															},
														},
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "*",
																},
															},
														},
													}},
												},
											},
										},
									},
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "PRIMARY"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "PRIMARY",
						Line:           12,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "OPTIONALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "REPEATEDSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "SPECIALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "GROUPEDSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "METAIDENTIFIER",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "TERMINAL"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "EMPTY"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "EMPTY",
						Line:           16,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "OPTIONALSEQUENCE",
						Line:           17,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "(/"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "/)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "REPEATEDSEQUENCE",
						Line:           18,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "(:"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: ":)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "GROUPEDSEQUENCE",
						Line:           19,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "("},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: ")"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "TERMINAL",
						Line:           20,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "'"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "CHARACTER"},
									},
									Exception: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "'"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "CHARACTER",
																},
															},
															Exception: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "'",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "'"},
									},
								},
							}},
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "\""},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "CHARACTER"},
									},
									Exception: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "\""},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "CHARACTER",
																},
															},
															Exception: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "\"",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "\""},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "METAIDENTIFIER",
						Line:           25,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "LETTER"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "LETTER",
																},
															},
														},
													},
												},
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "DIGIT",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "INTEGER",
						Line:           26,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "DIGIT"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "DIGIT",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "SPECIALSEQUENCE",
						Line:           27,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "?"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Exception: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	Terminal: "?",
																},
															},
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "CHARACTER",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "?"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "COMMENT",
						Line:           28,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "(*"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "COMMENTSYMBOL",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Terminal: "*)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "COMMENTSYMBOL",
						Line:           29,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "COMMENT"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "TERMINAL"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "SPECIALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "CHARACTER"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "EBNF definition part 1",
			grammar: `
(*
The syntax of Extended BNF can be defined using
itself. There are four parts in this example,
the first part names the characters, the second
part defines the removal of unnecessary non-
printing characters, the third part defines the
removal of textual comments, and the final part
defines the structure of Extended BNF itself.

Each syntax rule in this example starts with a
comment that identifies the corresponding clause
in the standard.

The meaning of special-sequences is not defined
in the standard. In this example (see the
reference to 7.6) they represent control
functions defined by ISO/IEC 6429:1992.
Another special-sequence defines a
syntactic-exception (see the reference to 4.7).
*)
(*
The first part of the lexical syntax defines the
characters in the 7-bit character set (ISO/IEC
646:1991) that represent each terminal-character
and gap-separator in Extended BNF.
*)
(* see 7.2 *) letter
= 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h'
| 'i'| 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p'
| 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x'
| 'y' | 'z'
| 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H'
| 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P'
| 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X'
| 'Y' | 'Z';
(* see 7.2 *) decimal digit
= '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7'
| '8' | '9';
(*
The representation of the following
terminal-characters is defined in clauses 7.3,
7.4 and tables 1, 2.
*)
concatenate symbol = ',';
defining symbol = '=';
definition separator symbol = '|' | '/' | '!';
end comment symbol = '*)';
end group symbol = ')';
end option symbol = ']' | '/)';
end repeat symbol = '}' | ':)';
except symbol = '-';
first quote symbol = "'";
repetition symbol = '*';
second quote symbol = '"';
special sequence symbol = '?';
start comment symbol = '(*';
start group symbol = '(';
start option symbol = '[' | '(/';
start repeat symbol = '{' | '(:';
terminator symbol = ';' | '.';
(* see 7.5 *) other character
= ' ' | ';' | '+' | '_' | '%' | '@'
| '&' | '#' | '$' | '<' | '>' | '\'
| '^' | '` + "`" + `' | '~';
(* see 7.6 *) space character = ' ';
horizontal tabulation character
= ? IS0 6429 character Horizontal Tabulation ? ;
new line
= { ? IS0 6429 character Carriage Return ? },
? IS0 6429 character Line Feed ?,
{ ? IS0 6429 character Carriage Return ? };
vertical tabulation character
= ? IS0 6429 character Vertical Tabulation ? ;
form feed
= ? IS0 6429 character Form Feed ? ;
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "letter",
						Line:           28,
						Comments: []string{
							`The syntax of Extended BNF can be defined using
itself. There are four parts in this example,
the first part names the characters, the second
part defines the removal of unnecessary non-
printing characters, the third part defines the
removal of textual comments, and the final part
defines the structure of Extended BNF itself.

Each syntax rule in this example starts with a
comment that identifies the corresponding clause
in the standard.

The meaning of special-sequences is not defined
in the standard. In this example (see the
reference to 7.6) they represent control
functions defined by ISO/IEC 6429:1992.
Another special-sequence defines a
syntactic-exception (see the reference to 4.7).`,
							`The first part of the lexical syntax defines the
characters in the 7-bit character set (ISO/IEC
646:1991) that represent each terminal-character
and gap-separator in Extended BNF.`,
							"see 7.2",
						},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "a"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "b"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "c"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "d"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "e"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "f"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "g"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "h"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "i"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "j"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "k"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "l"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "m"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "n"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "o"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "p"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "q"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "r"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "s"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "t"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "u"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "v"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "w"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "x"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "y"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "z"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "A"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "B"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "C"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "D"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "E"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "F"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "G"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "H"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "I"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "J"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "K"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "L"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "M"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "N"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "O"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "P"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "Q"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "R"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "S"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "T"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "U"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "V"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "W"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "X"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "Y"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "Z"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "decimaldigit",
						Line:           37,
						Comments:       []string{"see 7.2"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "0"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "1"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "2"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "3"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "4"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "5"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "6"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "7"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "8"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "9"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "concatenatesymbol",
						Line:           45,
						Comments: []string{
							`The representation of the following
terminal-characters is defined in clauses 7.3,
7.4 and tables 1, 2.`,
						},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ","},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "definingsymbol",
						Line:           46,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "="},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "definitionseparatorsymbol",
						Line:           47,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "|"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "/"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "!"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endcommentsymbol",
						Line:           48,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "*)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endgroupsymbol",
						Line:           49,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ")"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endoptionsymbol",
						Line:           50,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "]"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "/)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endrepeatsymbol",
						Line:           51,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "}"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ":)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "exceptsymbol",
						Line:           52,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "-"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "firstquotesymbol",
						Line:           53,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "'"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "repetitionsymbol",
						Line:           54,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "*"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "secondquotesymbol",
						Line:           55,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "\""},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "specialsequencesymbol",
						Line:           56,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "?"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startcommentsymbol",
						Line:           57,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "(*"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startgroupsymbol",
						Line:           58,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "("},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startoptionsymbol",
						Line:           59,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "["},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "(/"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startrepeatsymbol",
						Line:           60,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "{"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "(:"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "terminatorsymbol",
						Line:           61,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ";"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "."},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "othercharacter",
						Line:           62,
						Comments:       []string{"see 7.5"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: " "},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ";"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "+"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "_"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "%"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "@"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "&"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "#"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "$"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "<"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: ">"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "\\"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "^"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "`"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "~"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "spacecharacter",
						Line:           66,
						Comments:       []string{"see 7.6"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: " "},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "horizontaltabulationcharacter",
						Line:           67,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												SpecialSequence: "IS0 6429 character Horizontal Tabulation",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "newline",
						Line:           69,
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	SpecialSequence: "IS0 6429 character Carriage Return",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											SpecialSequence: "IS0 6429 character Line Feed",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	SpecialSequence: "IS0 6429 character Carriage Return",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "verticaltabulationcharacter",
						Line:           73,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												SpecialSequence: "IS0 6429 character Vertical Tabulation",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "formfeed",
						Line:           75,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												SpecialSequence: "IS0 6429 character Form Feed",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "EBNF definition part 2",
			grammar: `
(*
The second part of the syntax defines the
removal of unnecessary non-printing characters
from a syntax.
*)
(* see 6.2 *) terminal character
= letter
| decimal digit
| concatenate symbol
| defining symbol
| definition separator symbol
| end comment symbol
| end group symbol
| end option symbol
| end repeat symbol
| except symbol
| first quote symbol
| repetition symbol
| second quote symbol
| special sequence symbol
| start comment symbol
| start group symbol
| start option symbol
| start repeat symbol
| terminator symbol
| other character;
(* see 6.3 *) gap free symbol
= terminal character
- (first quote symbol | second quote symbol)
| terminal string;
(* see 4.16 *) terminal string
= first quote symbol, first terminal character,
{first terminal character},
first quote symbol
| second quote symbol, second terminal character,
{second terminal character},
second quote symbol;
(* see 4.17 *) first terminal character
= terminal character - first quote symbol;
(* see 4.18 *) second terminal character
= terminal character - second quote symbol;
(* see 6.4 *) gap separator
= space character
| horizontal tabulation character
| new line
| vertical tabulation character
| form feed;
(* see 6.5 *) syntax
= {gap separator},
gap free symbol, {gap separator},
{gap free symbol, {gap separator}};
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "terminalcharacter",
						Line:           7,
						Comments: []string{
							`The second part of the syntax defines the
removal of unnecessary non-printing characters
from a syntax.`,
							"see 6.2",
						},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "letter"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "decimaldigit",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "concatenatesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "definingsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "definitionseparatorsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "endcommentsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "endgroupsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "endoptionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "endrepeatsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "exceptsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "firstquotesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "repetitionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "secondquotesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "specialsequencesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "startcommentsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "startgroupsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "startoptionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "startrepeatsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminatorsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "othercharacter",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "gapfreesymbol",
						Line:           28,
						Comments:       []string{"see 6.3"},
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "terminalcharacter",
										},
									},
									Exception: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{GroupedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "firstquotesymbol",
															},
														},
													},
												},
											},
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "secondquotesymbol",
															},
														},
													},
												},
											},
										}},
									},
								},
							}},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminalstring",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "terminalstring",
						Line:           32,
						Comments:       []string{"see 4.16"},
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "firstquotesymbol",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "firstterminalcharacter",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "firstterminalcharacter",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "firstquotesymbol",
										},
									},
								},
							}},
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "secondquotesymbol",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "secondterminalcharacter",
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "secondterminalcharacter",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "secondquotesymbol",
										},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "firstterminalcharacter",
						Line:           39,
						Comments:       []string{"see 4.17"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "firstquotesymbol",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "secondterminalcharacter",
						Line:           41,
						Comments:       []string{"see 4.18"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "secondquotesymbol",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "gapseparator",
						Line:           43,
						Comments:       []string{"see 6.4"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "spacecharacter",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "horizontaltabulationcharacter",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "newline"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "verticaltabulationcharacter",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "formfeed"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "syntax",
						Line:           49,
						Comments:       []string{"see 6.5"},
						Definitions: iso.DefinitionsList{{Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "gapseparator",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{MetaIdentifier: "gapfreesymbol"},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "gapseparator",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{{Terms: []iso.Term{
											{
												Factor: iso.Factor{
													Repetitions: -1,
													Primary: iso.Primary{
														MetaIdentifier: "gapfreesymbol",
													},
												},
											},
											{
												Factor: iso.Factor{
													Repetitions: -1,
													Primary: iso.Primary{
														RepeatedSequence: iso.DefinitionsList{
															{
																Terms: []iso.Term{
																	{
																		Factor: iso.Factor{
																			Repetitions: -1,
																			Primary: iso.Primary{
																				MetaIdentifier: "gapseparator",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										}}},
									},
								},
							},
						}}},
					},
				},
			},
		},
		{
			name: "EBNF definition part 3",
			grammar: `
(*
The third part of the syntax defines the
removal of bracketed-textual-comments from
gap-free-symbols that form a syntax.
*)
(* see 6.6 *) commentless symbol
= terminal character
- (letter
| decimal digit
| first quote symbol
| second quote symbol
| start comment symbol
| end comment symbol
| special sequence symbol
| other character)
| meta identifier
| integer
| terminal string
| special sequence;
(* see 4.9 *) integer
= decimal digit, {decimal digit};
(* see 4.14 *) meta identifier
= letter, {meta identifier character};
(* see 4.15 *) meta identifier character
= letter
| decimal digit;
(* see 4.19 *) special sequence
= special sequence symbol,
{special sequence character},
special sequence symbol;
(* see 4.20 *) special sequence character
= terminal character - special sequence symbol;
(* see 6.7 *) comment symbol
= bracketed textual comment
| other character
| commentless symbol;
(* see 6.8 *) bracketed textual comment
= start comment symbol, {comment symbol},
end comment symbol;
(* see 6.9 *) syntax
= {bracketed textual comment},
commentless symbol,
{bracketed textual comment},
{commentless symbol,
{bracketed textual comment}};
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "commentlesssymbol",
						Line:           7,
						Comments: []string{
							`The third part of the syntax defines the
removal of bracketed-textual-comments from
gap-free-symbols that form a syntax.`,
							"see 6.6",
						},
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{MetaIdentifier: "terminalcharacter"},
								},
								Exception: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{GroupedSequence: iso.DefinitionsList{
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "letter",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "decimaldigit",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "firstquotesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "secondquotesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "startcommentsymbol",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "endcommentsymbol",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "specialsequencesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "othercharacter",
														},
													},
												},
											},
										},
									}},
								},
							}}},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "metaidentifier",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "integer"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminalstring",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "specialsequence",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "integer",
						Line:           21,
						Comments:       []string{"see 4.9"},
						Definitions: iso.DefinitionsList{
							{Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "decimaldigit"},
									},
								},
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											RepeatedSequence: iso.DefinitionsList{
												{
													Terms: []iso.Term{
														{
															Factor: iso.Factor{
																Repetitions: -1,
																Primary: iso.Primary{
																	MetaIdentifier: "decimaldigit",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "metaidentifier",
						Line:           23,
						Comments:       []string{"see 4.14"},
						Definitions: iso.DefinitionsList{{Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{MetaIdentifier: "letter"},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "metaidentifiercharacter",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						}}},
					},
					{
						MetaIdentifier: "metaidentifiercharacter",
						Line:           25,
						Comments:       []string{"see 4.15"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{MetaIdentifier: "letter"},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "decimaldigit",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "specialsequence",
						Line:           28,
						Comments:       []string{"see 4.19"},
						Definitions: iso.DefinitionsList{{Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										MetaIdentifier: "specialsequencesymbol",
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "specialsequencecharacter",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										MetaIdentifier: "specialsequencesymbol",
									},
								},
							},
						}}},
					},
					{
						MetaIdentifier: "specialsequencecharacter",
						Line:           32,
						Comments:       []string{"see 4.20"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "specialsequencesymbol",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "commentsymbol",
						Line:           34,
						Comments:       []string{"see 6.7"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "bracketedtextualcomment",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "othercharacter",
											},
										},
									},
								},
							},
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary: iso.Primary{
												MetaIdentifier: "commentlesssymbol",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "bracketedtextualcomment",
						Line:           38,
						Comments:       []string{"see 6.8"},
						Definitions: iso.DefinitionsList{{Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										MetaIdentifier: "startcommentsymbol",
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "commentsymbol",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{MetaIdentifier: "endcommentsymbol"},
								},
							},
						}}},
					},
					{
						MetaIdentifier: "syntax",
						Line:           41,
						Comments:       []string{"see 6.9"},
						Definitions: iso.DefinitionsList{{Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "bracketedtextualcomment",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{MetaIdentifier: "commentlesssymbol"},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{
											{
												Terms: []iso.Term{
													{
														Factor: iso.Factor{
															Repetitions: -1,
															Primary: iso.Primary{
																MetaIdentifier: "bracketedtextualcomment",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary: iso.Primary{
										RepeatedSequence: iso.DefinitionsList{{Terms: []iso.Term{
											{
												Factor: iso.Factor{
													Repetitions: -1,
													Primary: iso.Primary{
														MetaIdentifier: "commentlesssymbol",
													},
												},
											},
											{
												Factor: iso.Factor{
													Repetitions: -1,
													Primary: iso.Primary{
														RepeatedSequence: iso.DefinitionsList{
															{
																Terms: []iso.Term{
																	{
																		Factor: iso.Factor{
																			Repetitions: -1,
																			Primary: iso.Primary{
																				MetaIdentifier: "bracketedtextualcomment",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										}}},
									},
								},
							},
						}}},
					},
				},
			},
		},
		{
			name: "EBNF definition part 4",
			grammar: `
(*
The final part of the syntax defines the
abstract syntax of Extended BNF, i.e. the
structure in terms of the commentless symbols.
*)
(* see 4.2 *) syntax
= syntax rule, {syntax rule};
(* see 4.3 *) syntax rule
= meta identifier, defining symbol,
definitions list, terminator symbol;
(* see 4.4 *) definitions list
= single definition,
{definition separator symbol,
single definition};
(* see 4.5 *) single definition
= syntactic term,
{concatenate symbol, syntactic term};
(* see 4.6 *) syntactic term
= syntactic factor,
[except symbol, syntactic exception];
(* see 4.7 *) syntactic exception
= ? a syntactic-factor that could be replaced
by a syntactic-factor containing no
meta-identifiers
? ;
(* see 4.8 *) syntactic factor
= [integer, repetition symbol],
syntactic primary;
(* see 4.10 *) syntactic primary
= optional sequence
| repeated sequence
| grouped sequence
| meta identifier
| terminal string
| special sequence
| empty sequence;
(* see 4.11 *) optional sequence
= start option symbol, definitions list,
end option symbol;
(* see 4.12 *) repeated sequence
= start repeat symbol, definitions list,
end repeat symbol;
(* see 4.13 *) grouped sequence
= start group symbol, definitions list,
end group symbol;
(* see 4.21 *) empty sequence
= ;
`,
			expectedSyntax: iso.Syntax{Rules: []iso.Rule{
				{
					MetaIdentifier: "syntax",
					Line:           7,
					Comments: []string{
						`The final part of the syntax defines the
abstract syntax of Extended BNF, i.e. the
structure in terms of the commentless symbols.`,
						"see 4.2",
					},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "syntaxrule"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary: iso.Primary{
									RepeatedSequence: iso.DefinitionsList{
										{
											Terms: []iso.Term{
												{
													Factor: iso.Factor{
														Repetitions: -1,
														Primary: iso.Primary{
															MetaIdentifier: "syntaxrule",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "syntaxrule",
					Line:           9,
					Comments:       []string{"see 4.3"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "metaidentifier"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "definingsymbol"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "terminatorsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "definitionslist",
					Line:           12,
					Comments:       []string{"see 4.4"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "singledefinition"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary: iso.Primary{
									RepeatedSequence: iso.DefinitionsList{{Terms: []iso.Term{
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "definitionseparatorsymbol",
												},
											},
										},
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "singledefinition",
												},
											},
										},
									}}},
								},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "singledefinition",
					Line:           16,
					Comments:       []string{"see 4.5"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "syntacticterm"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary: iso.Primary{
									RepeatedSequence: iso.DefinitionsList{{Terms: []iso.Term{
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "concatenatesymbol",
												},
											},
										},
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "syntacticterm",
												},
											},
										},
									}}},
								},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "syntacticterm",
					Line:           19,
					Comments:       []string{"see 4.6"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "syntacticfactor"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary: iso.Primary{
									OptionalSequence: iso.DefinitionsList{{Terms: []iso.Term{
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "exceptsymbol",
												},
											},
										},
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "syntacticexception",
												},
											},
										},
									}}},
								},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "syntacticexception",
					Line:           22,
					Comments:       []string{"see 4.7"},
					Definitions: iso.DefinitionsList{
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											SpecialSequence: `a syntactic-factor that could be replaced
by a syntactic-factor containing no
meta-identifiers`,
										},
									},
								},
							},
						},
					},
				},
				{
					MetaIdentifier: "syntacticfactor",
					Line:           27,
					Comments:       []string{"see 4.8"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary: iso.Primary{
									OptionalSequence: iso.DefinitionsList{{Terms: []iso.Term{
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "integer",
												},
											},
										},
										{
											Factor: iso.Factor{
												Repetitions: -1,
												Primary: iso.Primary{
													MetaIdentifier: "repetitionsymbol",
												},
											},
										},
									}}},
								},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "syntacticprimary"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "syntacticprimary",
					Line:           30,
					Comments:       []string{"see 4.10"},
					Definitions: iso.DefinitionsList{
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "optionalsequence",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "repeatedsequence",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "groupedsequence",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "metaidentifier",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "terminalstring",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary: iso.Primary{
											MetaIdentifier: "specialsequence",
										},
									},
								},
							},
						},
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{MetaIdentifier: "emptysequence"},
									},
								},
							},
						},
					},
				},
				{
					MetaIdentifier: "optionalsequence",
					Line:           38,
					Comments:       []string{"see 4.11"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "startoptionsymbol"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "endoptionsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "repeatedsequence",
					Line:           41,
					Comments:       []string{"see 4.12"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "startrepeatsymbol"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "endrepeatsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "groupedsequence",
					Line:           44,
					Comments:       []string{"see 4.13"},
					Definitions: iso.DefinitionsList{{Terms: []iso.Term{
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "startgroupsymbol"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: iso.Factor{
								Repetitions: -1,
								Primary:     iso.Primary{MetaIdentifier: "endgroupsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "emptysequence",
					Line:           47,
					Comments:       []string{"see 4.21"},
					Definitions: iso.DefinitionsList{
						{
							Terms: []iso.Term{
								{
									Factor: iso.Factor{
										Repetitions: -1,
										Primary:     iso.Primary{Empty: true},
									},
								},
							},
						},
					},
				},
			}},
		},
		{
			name: "Syntax with trailing comments",
			grammar: `
a = ;
(* a trailing comment *)`,
			expectedSyntax: iso.Syntax{
				TrailingComments: []string{"a trailing comment"},
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Line:           2,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Syntax with comment before defining symbol",
			grammar: `a (* comment before defining symbol*) = ;`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Comments:       []string{"comment before defining symbol"},
						Line:           1,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Syntax with comments on factor",
			grammar: `a = (* comment on factor *) ;`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Line:           1,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Comments:    []string{"comment on factor"},
											Repetitions: -1,
											Primary:     iso.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Syntax with repetitions",
			grammar: `a = 1 * "a" ;`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Line:           1,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: 1,
											Primary:     iso.Primary{Terminal: "a"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Syntax with comment after factor repetitions",
			grammar: `
a = 1 * (* comment after repetitions symbol *) "a" ;
b = 2 (* comment after repetitions integer *) * "b" ;
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Line:           2,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Comments: []string{
												"comment after repetitions symbol",
											},
											Repetitions: 1,
											Primary:     iso.Primary{Terminal: "a"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "b",
						Line:           3,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Comments: []string{
												"comment after repetitions integer",
											},
											Repetitions: 2,
											Primary:     iso.Primary{Terminal: "b"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Syntax with newlines in terminal strings",
			grammar: `
(* terminal string beginning with newline *) a = "
stuff" ;
(* terminal string containing newline *) b = "word
second" ;
c = ;
`,
			expectedSyntax: iso.Syntax{
				Rules: []iso.Rule{
					{
						MetaIdentifier: "a",
						Line:           2,
						Comments:       []string{"terminal string beginning with newline"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "\nstuff"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "b",
						Line:           4,
						Comments:       []string{"terminal string containing newline"},
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Terminal: "word\nsecond"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "c",
						Line:           6,
						Definitions: iso.DefinitionsList{
							{
								Terms: []iso.Term{
									{
										Factor: iso.Factor{
											Repetitions: -1,
											Primary:     iso.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Syntax with comments containing other symbols",
			grammar: `
(*
(* "double quoted terminal" 'single quoted terminal' ? special sequence ? *)
*)
a = ;`,
			expectedSyntax: iso.Syntax{Rules: []iso.Rule{{
				Comments: []string{
					"(* \"double quoted terminal\" 'single quoted terminal' ? special sequence ? *)",
				},
				Line:           5,
				MetaIdentifier: "a",
				Definitions: iso.DefinitionsList{
					{
						Terms: []iso.Term{
							{
								Factor: iso.Factor{
									Repetitions: -1,
									Primary:     iso.Primary{Empty: true},
								},
							},
						},
					},
				},
			}}},
		},
	}

	for _, tc := range tcs {
		parser := iso.New()
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			syntax, err := parser.Parse(tc.grammar)
			if err != nil {
				t.Fatalf("Got unexpected error %s.", err)
			}
			assertSyntaxesEqual(t, tc.expectedSyntax, syntax)
		})
	}
}
