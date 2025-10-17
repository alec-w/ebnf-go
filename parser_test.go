package ebnf_test

import (
	"testing"

	"github.com/alec-w/ebnf-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
				cases.Title(language.BritishEnglish).String(itemSingular),
				i+1,
			)
			t.Fail()
			failed = true
		}
	}
	return !failed
}

func assertSyntaxesEqual(t *testing.T, expected, actual ebnf.Syntax) bool {
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

func assertRulesEqual(t *testing.T, expected, actual ebnf.Rule) bool {
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
	if !assertSlicesEqual(
		t,
		expected.Comments,
		actual.Comments,
		"comments",
		"comment",
		assertCommentsEqual,
	) {
		if expected.MetaIdentifier != "" && expected.MetaIdentifier == actual.MetaIdentifier {
			t.Logf("Rule %q comments were not equal", expected.MetaIdentifier)
		} else {
			t.Log("Rule comments were not equal")
		}
		t.Fail()
		failed = true
	}
	if expected.Line != actual.Line {
		if expected.MetaIdentifier != "" && expected.MetaIdentifier == actual.MetaIdentifier {
			t.Logf(
				"Expected rule %q to be on line %d. Got %d.",
				expected.MetaIdentifier,
				expected.Line,
				actual.Line,
			)
		} else {
			t.Logf("Expected rule to be on line %d. Got %d.", expected.Line, actual.Line)
		}
		t.Fail()
		failed = true
	}
	if assertSlicesEqual(
		t,
		expected.Definitions,
		actual.Definitions,
		"definitions",
		"definition",
		assertDefinitionsEqual,
	) {
		return !failed
	}
	if expected.MetaIdentifier != "" && expected.MetaIdentifier == actual.MetaIdentifier {
		t.Logf("Rule %q definitions were not equal", expected.MetaIdentifier)
	} else {
		t.Log("Rule definitions were not equal")
	}
	t.Fail()
	return false
}

func assertDefinitionsEqual(t *testing.T, expected, actual ebnf.Definition) bool {
	t.Helper()
	if assertSlicesEqual(t, expected.Terms, actual.Terms, "terms", "term", assertTermsEqual) {
		return true
	}
	t.Log("Definition terms were not equal")
	t.Fail()
	return false
}

func assertTermsEqual(t *testing.T, expected, actual ebnf.Term) bool {
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

func assertFactorsEqual(t *testing.T, expected, actual ebnf.Factor) bool {
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
	switch {
	case expected.Primary.OptionalSequence != nil:
		if !assertSlicesEqual(
			t,
			expected.Primary.OptionalSequence,
			actual.Primary.OptionalSequence,
			"definitions",
			"definition",
			assertDefinitionsEqual,
		) {
			t.Log("Factor primary optional sequences not equal")
			t.Fail()
			failed = true
		}
	case expected.Primary.RepeatedSequence != nil:
		if !assertSlicesEqual(
			t,
			expected.Primary.RepeatedSequence,
			actual.Primary.RepeatedSequence,
			"definitions",
			"definition",
			assertDefinitionsEqual,
		) {
			t.Log("Factor primary repeated sequences not equal")
			t.Fail()
			failed = true
		}
	case expected.Primary.SpecialSequence != "":
		if expected.Primary.SpecialSequence != actual.Primary.SpecialSequence {
			t.Logf(
				"Expected factor primary special sequence %q. Got %q.",
				expected.Primary.SpecialSequence,
				actual.Primary.SpecialSequence,
			)
			t.Fail()
			failed = true
		}
	case expected.Primary.GroupedSequence != nil:
		if !assertSlicesEqual(
			t,
			expected.Primary.GroupedSequence,
			actual.Primary.GroupedSequence,
			"definitions",
			"definition",
			assertDefinitionsEqual,
		) {
			t.Log("Factor primary grouped sequences not equal")
			t.Fail()
			failed = true
		}
	case expected.Primary.MetaIdentifier != "":
		if expected.Primary.MetaIdentifier != actual.Primary.MetaIdentifier {
			t.Logf(
				"Expected factor primary meta identifier %q. Got %q.",
				expected.Primary.MetaIdentifier,
				actual.Primary.MetaIdentifier,
			)
			t.Fail()
			failed = true
		}
	case expected.Primary.Terminal != "":
		if expected.Primary.Terminal != actual.Primary.Terminal {
			t.Logf(
				"Expected factor primary terminal %q. Got %q.",
				expected.Primary.Terminal,
				actual.Primary.Terminal,
			)
			t.Fail()
			failed = true
		}
	case expected.Primary.Empty:
		if expected.Primary.Empty != actual.Primary.Empty {
			t.Logf(
				"Expected factor primary empty %t. Got %t.",
				expected.Primary.Empty,
				actual.Primary.Empty,
			)
			t.Fail()
			failed = true
		}
	default:
	}
	return !failed
}

func TestParseSyntax(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name           string
		grammar        string
		expectedSyntax ebnf.Syntax
	}{
		{
			name: "Positive integer definition",
			grammar: `
nonZeroDigit = "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
digit = "0" | nonZeroDigit ;
integer = "0" | nonZeroDigit, { digit } ;
`,
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
					{
						MetaIdentifier: "nonZeroDigit",
						Line:           2,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "1"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "2"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "3"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "4"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "5"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "6"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "7"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "8"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "9"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "digit",
						Line:           3,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "0"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "0"},
										},
									},
								},
							},
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "nonZeroDigit"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
					{
						MetaIdentifier: "SYNTAX",
						Line:           2,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "SYNTAXRULE",
											},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												RepeatedSequence: ebnf.DefinitionsList{
													{
														Terms: []ebnf.Term{
															{
																Factor: ebnf.Factor{
																	Repetitions: -1,
																	Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "METAIDENTIFIER",
											},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "="},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "DEFINITIONSLIST",
											},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "."},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "DEFINITIONSLIST",
						Line:           5,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "SINGLEDEFINITION",
											},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												RepeatedSequence: ebnf.DefinitionsList{
													{Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	Terminal: "/",
																},
															},
														},
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "TERM"},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												RepeatedSequence: ebnf.DefinitionsList{
													{Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	Terminal: ",",
																},
															},
														},
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "FACTOR"},
										},
									},
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												OptionalSequence: ebnf.DefinitionsList{
													{Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	Terminal: "-",
																},
															},
														},
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "FACTOR"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "FACTOR",
						Line:           11,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												OptionalSequence: ebnf.DefinitionsList{
													{Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	MetaIdentifier: "INTEGER",
																},
															},
														},
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "PRIMARY"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "PRIMARY",
						Line:           12,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "OPTIONALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "REPEATEDSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "SPECIALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "GROUPEDSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "METAIDENTIFIER",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "TERMINAL"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "EMPTY"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "EMPTY",
						Line:           16,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Empty: true},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "OPTIONALSEQUENCE",
						Line:           17,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "(/"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "/)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "REPEATEDSEQUENCE",
						Line:           18,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "(:"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: ":)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "GROUPEDSEQUENCE",
						Line:           19,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "("},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "DEFINITIONSLIST",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: ")"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "TERMINAL",
						Line:           20,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "'"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "CHARACTER"},
									},
									Exception: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "'"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	MetaIdentifier: "CHARACTER",
																},
															},
															Exception: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "'"},
									},
								},
							}},
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "\""},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "CHARACTER"},
									},
									Exception: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "\""},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	MetaIdentifier: "CHARACTER",
																},
															},
															Exception: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "\""},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "METAIDENTIFIER",
						Line:           25,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "LETTER"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	MetaIdentifier: "LETTER",
																},
															},
														},
													},
												},
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "DIGIT"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "?"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Exception: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
																	Terminal: "?",
																},
															},
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "?"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "COMMENT",
						Line:           28,
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "(*"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Terminal: "*)"},
									},
								},
							}},
						},
					},
					{
						MetaIdentifier: "COMMENTSYMBOL",
						Line:           29,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "COMMENT"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "TERMINAL"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "SPECIALSEQUENCE",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "CHARACTER"},
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
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "a"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "b"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "c"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "d"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "e"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "f"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "g"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "h"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "i"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "j"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "k"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "l"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "m"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "n"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "o"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "p"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "q"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "r"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "s"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "t"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "u"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "v"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "w"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "x"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "y"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "z"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "A"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "B"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "C"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "D"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "E"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "F"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "G"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "H"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "I"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "J"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "K"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "L"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "M"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "N"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "O"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "P"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "Q"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "R"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "S"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "T"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "U"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "V"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "W"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "X"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "Y"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "Z"},
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "0"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "1"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "2"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "3"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "4"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "5"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "6"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "7"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "8"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "9"},
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ","},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "definingsymbol",
						Line:           46,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "="},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "definitionseparatorsymbol",
						Line:           47,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "|"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "/"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "!"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endcommentsymbol",
						Line:           48,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "*)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endgroupsymbol",
						Line:           49,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ")"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endoptionsymbol",
						Line:           50,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "]"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "/)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "endrepeatsymbol",
						Line:           51,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "}"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ":)"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "exceptsymbol",
						Line:           52,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "-"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "firstquotesymbol",
						Line:           53,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "'"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "repetitionsymbol",
						Line:           54,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "*"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "secondquotesymbol",
						Line:           55,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "\""},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "specialsequencesymbol",
						Line:           56,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "?"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startcommentsymbol",
						Line:           57,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "(*"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startgroupsymbol",
						Line:           58,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "("},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startoptionsymbol",
						Line:           59,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "["},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "(/"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "startrepeatsymbol",
						Line:           60,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "{"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "(:"},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "terminatorsymbol",
						Line:           61,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ";"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "."},
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: " "},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ";"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "+"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "_"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "%"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "@"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "&"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "#"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "$"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "<"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: ">"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "\\"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "^"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "`"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: "~"},
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{Terminal: " "},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "horizontaltabulationcharacter",
						Line:           67,
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											SpecialSequence: "IS0 6429 character Line Feed",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
					{
						MetaIdentifier: "terminalcharacter",
						Line:           7,
						Comments: []string{
							`The second part of the syntax defines the
removal of unnecessary non-printing characters
from a syntax.`,
							"see 6.2",
						},
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "letter"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "decimaldigit",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "concatenatesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "definingsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "definitionseparatorsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "endcommentsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "endgroupsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "endoptionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "endrepeatsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "exceptsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "firstquotesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "repetitionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "secondquotesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "specialsequencesymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "startcommentsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "startgroupsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "startoptionsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "startrepeatsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminatorsymbol",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "terminalcharacter",
										},
									},
									Exception: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{GroupedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
																MetaIdentifier: "firstquotesymbol",
															},
														},
													},
												},
											},
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "firstquotesymbol",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "firstterminalcharacter",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "firstquotesymbol",
										},
									},
								},
							}},
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "secondquotesymbol",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "secondterminalcharacter",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "spacecharacter",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "horizontaltabulationcharacter",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "newline"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "verticaltabulationcharacter",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "formfeed"},
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
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "gapfreesymbol"},
								},
							},
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
											{
												Factor: ebnf.Factor{
													Repetitions: -1,
													Primary: ebnf.Primary{
														MetaIdentifier: "gapfreesymbol",
													},
												},
											},
											{
												Factor: ebnf.Factor{
													Repetitions: -1,
													Primary: ebnf.Primary{
														RepeatedSequence: ebnf.DefinitionsList{
															{
																Terms: []ebnf.Term{
																	{
																		Factor: ebnf.Factor{
																			Repetitions: -1,
																			Primary: ebnf.Primary{
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
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
					{
						MetaIdentifier: "commentlesssymbol",
						Line:           7,
						Comments: []string{
							`The third part of the syntax defines the
removal of bracketed-textual-comments from
gap-free-symbols that form a syntax.`,
							"see 6.6",
						},
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "terminalcharacter"},
								},
								Exception: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{GroupedSequence: ebnf.DefinitionsList{
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "letter",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "decimaldigit",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "firstquotesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "secondquotesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "startcommentsymbol",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "endcommentsymbol",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
															MetaIdentifier: "specialsequencesymbol",
														},
													},
												},
											},
										},
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
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
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "metaidentifier",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "integer"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminalstring",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "decimaldigit"},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											RepeatedSequence: ebnf.DefinitionsList{
												{
													Terms: []ebnf.Term{
														{
															Factor: ebnf.Factor{
																Repetitions: -1,
																Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "letter"},
								},
							},
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "letter"},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										MetaIdentifier: "specialsequencesymbol",
									},
								},
							},
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminalcharacter",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "bracketedtextualcomment",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "othercharacter",
											},
										},
									},
								},
							},
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
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
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										MetaIdentifier: "startcommentsymbol",
									},
								},
							},
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "endcommentsymbol"},
								},
							},
						}}},
					},
					{
						MetaIdentifier: "syntax",
						Line:           41,
						Comments:       []string{"see 6.9"},
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "commentlesssymbol"},
								},
							},
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{
											{
												Terms: []ebnf.Term{
													{
														Factor: ebnf.Factor{
															Repetitions: -1,
															Primary: ebnf.Primary{
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
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										RepeatedSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
											{
												Factor: ebnf.Factor{
													Repetitions: -1,
													Primary: ebnf.Primary{
														MetaIdentifier: "commentlesssymbol",
													},
												},
											},
											{
												Factor: ebnf.Factor{
													Repetitions: -1,
													Primary: ebnf.Primary{
														RepeatedSequence: ebnf.DefinitionsList{
															{
																Terms: []ebnf.Term{
																	{
																		Factor: ebnf.Factor{
																			Repetitions: -1,
																			Primary: ebnf.Primary{
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
			expectedSyntax: ebnf.Syntax{Rules: []ebnf.Rule{
				{
					MetaIdentifier: "syntax",
					Line:           7,
					Comments: []string{
						`The final part of the syntax defines the
abstract syntax of Extended BNF, i.e. the
structure in terms of the commentless symbols.`,
						"see 4.2",
					},
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "syntaxrule"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary: ebnf.Primary{
									RepeatedSequence: ebnf.DefinitionsList{
										{
											Terms: []ebnf.Term{
												{
													Factor: ebnf.Factor{
														Repetitions: -1,
														Primary: ebnf.Primary{
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
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "metaidentifier"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "definingsymbol"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "terminatorsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "definitionslist",
					Line:           12,
					Comments:       []string{"see 4.4"},
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "singledefinition"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary: ebnf.Primary{
									RepeatedSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
													MetaIdentifier: "definitionseparatorsymbol",
												},
											},
										},
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
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
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "syntacticterm"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary: ebnf.Primary{
									RepeatedSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
													MetaIdentifier: "concatenatesymbol",
												},
											},
										},
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
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
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "syntacticfactor"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary: ebnf.Primary{
									OptionalSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
													MetaIdentifier: "exceptsymbol",
												},
											},
										},
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
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
					Definitions: ebnf.DefinitionsList{
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
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
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary: ebnf.Primary{
									OptionalSequence: ebnf.DefinitionsList{{Terms: []ebnf.Term{
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
													MetaIdentifier: "integer",
												},
											},
										},
										{
											Factor: ebnf.Factor{
												Repetitions: -1,
												Primary: ebnf.Primary{
													MetaIdentifier: "repetitionsymbol",
												},
											},
										},
									}}},
								},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "syntacticprimary"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "syntacticprimary",
					Line:           30,
					Comments:       []string{"see 4.10"},
					Definitions: ebnf.DefinitionsList{
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "optionalsequence",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "repeatedsequence",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "groupedsequence",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "metaidentifier",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "terminalstring",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "specialsequence",
										},
									},
								},
							},
						},
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "emptysequence"},
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
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "startoptionsymbol"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "endoptionsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "repeatedsequence",
					Line:           41,
					Comments:       []string{"see 4.12"},
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "startrepeatsymbol"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "endrepeatsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "groupedsequence",
					Line:           44,
					Comments:       []string{"see 4.13"},
					Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "startgroupsymbol"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "definitionslist"},
							},
						},
						{
							Factor: ebnf.Factor{
								Repetitions: -1,
								Primary:     ebnf.Primary{MetaIdentifier: "endgroupsymbol"},
							},
						},
					}}},
				},
				{
					MetaIdentifier: "emptysequence",
					Line:           47,
					Comments:       []string{"see 4.21"},
					Definitions: ebnf.DefinitionsList{
						{
							Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{Empty: true},
									},
								},
							},
						},
					},
				},
			}},
		},
	}

	for _, tc := range tcs {
		parser := ebnf.New()
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
