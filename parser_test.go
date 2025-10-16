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
SYNTAX = SYNTAXRULE, (: SYNTAXRULE :).
SYNTAXRULE
= METAIDENTIFIER, '=', DEFINITIONSLIST, '.'.
DEFINITIONSLIST
= SINGLEDEFINITION,
(: '/', SINGLEDEFINITION :).
SINGLEDEFINITION = TERM, (: ',', TERM :).
TERM = FACTOR, (/ '-', EXCEPTION /).
EXCEPTION = FACTOR.
FACTOR = (/ INTEGER, '*' /), PRIMARY.
PRIMARY
= OPTIONALSEQUENCE / REPEATEDSEQUENCE
/ SPECIALSEQUENCE / GROUPEDSEQUENCE
/ METAIDENTIFIER / TERMINAL / EMPTY.
EMPTY = .
OPTIONALSEQUENCE = '(/', DEFINITIONSLIST, '/)'.
REPEATEDSEQUENCE = '(:', DEFINITIONSLIST, ':)'.
GROUPEDSEQUENCE = '(', DEFINITIONSLIST, ')'.
TERMINAL
= "'" , CHARACTER - "'",
(: CHARACTER - "'" :), "'"
/ '"' , CHARACTER - '"',
(: CHARACTER - '"' :), '"'.
METAIDENTIFIER = LETTER, (: LETTER / DIGIT :).
INTEGER = DIGIT, (: DIGIT :).
SPECIALSEQUENCE = '?', (: CHARACTER - '?' :), '?'.
COMMENT = '(*', (: COMMENTSYMBOL :), '*)'.
COMMENTSYMBOL
= COMMENT / TERMINAL / SPECIALSEQUENCE
/ CHARACTER.`,
			expectedSyntax: ebnf.Syntax{
				Rules: []ebnf.Rule{
					{
						MetaIdentifier: "SYNTAX",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary:     ebnf.Primary{MetaIdentifier: "SYNTAXRULE"},
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
						Comments: []string{
							`
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
`, `
The first part of the lexical syntax defines the
characters in the 7-bit character set (ISO/IEC
646:1991) that represent each terminal-character
and gap-separator in Extended BNF.
`,
							` see 7.2 `,
						},
						MetaIdentifier: "letter",
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
						Comments:       []string{" see 7.2 "},
						MetaIdentifier: "decimal digit",
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
						Comments: []string{`
The representation of the following
terminal-characters is defined in clauses 7.3,
7.4 and tables 1, 2.
`},
						MetaIdentifier: "concatenate symbol",
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
						MetaIdentifier: "defining symbol",
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
						MetaIdentifier: "definition separator symbol",
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
						MetaIdentifier: "end comment symbol",
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
						MetaIdentifier: "end group symbol",
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
						MetaIdentifier: "end option symbol",
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
						MetaIdentifier: "end repeat symbol",
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
						MetaIdentifier: "except symbol",
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
						MetaIdentifier: "first quote symbol",
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
						MetaIdentifier: "repetition symbol",
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
						MetaIdentifier: "second quote symbol",
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
						MetaIdentifier: "special sequence symbol",
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
						MetaIdentifier: "start comment symbol",
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
						MetaIdentifier: "start group symbol",
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
						MetaIdentifier: "start option symbol",
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
						MetaIdentifier: "start repeat symbol",
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
						MetaIdentifier: "terminator symbol",
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
						Comments:       []string{" see 7.5 "},
						MetaIdentifier: "other character",
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
						Comments:       []string{" see 7.6 "},
						MetaIdentifier: "space character",
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
						MetaIdentifier: "horizontal tabulation character",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												SpecialSequence: " IS0 6429 character Horizontal Tabulation ",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "new line",
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
																	SpecialSequence: " IS0 6429 character Carriage Return ",
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
											SpecialSequence: " IS0 6429 character Line Feed ",
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
																	SpecialSequence: " IS0 6429 character Carriage Return ",
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
						MetaIdentifier: "vertical tabulation character",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												SpecialSequence: " IS0 6429 character Vertical Tabulation ",
											},
										},
									},
								},
							},
						},
					},
					{
						MetaIdentifier: "form feed",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												SpecialSequence: " IS0 6429 character Form Feed ",
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
						Comments: []string{
							`
The second part of the syntax defines the
removal of unnecessary non-printing characters
from a syntax.
`,
							" see 6.2 ",
						},
						MetaIdentifier: "terminal character",
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
												MetaIdentifier: "decimal digit",
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
												MetaIdentifier: "concatenate symbol",
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
												MetaIdentifier: "defining symbol",
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
												MetaIdentifier: "definition separator symbol",
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
												MetaIdentifier: "end comment symbol",
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
												MetaIdentifier: "end group symbol",
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
												MetaIdentifier: "end option symbol",
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
												MetaIdentifier: "end repeat symbol",
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
												MetaIdentifier: "except symbol",
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
												MetaIdentifier: "first quote symbol",
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
												MetaIdentifier: "repetition symbol",
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
												MetaIdentifier: "second quote symbol",
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
												MetaIdentifier: "special sequence symbol",
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
												MetaIdentifier: "start comment symbol",
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
												MetaIdentifier: "start group symbol",
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
												MetaIdentifier: "start option symbol",
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
												MetaIdentifier: "start repeat symbol",
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
												MetaIdentifier: "terminator symbol",
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
												MetaIdentifier: "other character",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 6.3 "},
						MetaIdentifier: "gap free symbol",
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "terminal character",
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
																MetaIdentifier: "first quote symbol",
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
																MetaIdentifier: "second quote symbol",
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
												MetaIdentifier: "terminal string",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 4.16 "},
						MetaIdentifier: "terminal string",
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "first quote symbol",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "first terminal character",
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
																	MetaIdentifier: "first terminal character",
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
											MetaIdentifier: "first quote symbol",
										},
									},
								},
							}},
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "second quote symbol",
										},
									},
								},
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary: ebnf.Primary{
											MetaIdentifier: "second terminal character",
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
																	MetaIdentifier: "second terminal character",
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
											MetaIdentifier: "second quote symbol",
										},
									},
								},
							}},
						},
					},
					{
						Comments:       []string{" see 4.17 "},
						MetaIdentifier: "first terminal character",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminal character",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "first quote symbol",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 4.18 "},
						MetaIdentifier: "second terminal character",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminal character",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "second quote symbol",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 6.4 "},
						MetaIdentifier: "gap separator",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "space character",
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
												MetaIdentifier: "horizontal tabulation character",
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
											Primary:     ebnf.Primary{MetaIdentifier: "new line"},
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
												MetaIdentifier: "vertical tabulation character",
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
											Primary:     ebnf.Primary{MetaIdentifier: "form feed"},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 6.5 "},
						MetaIdentifier: "syntax",
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
																MetaIdentifier: "gap separator",
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
									Primary:     ebnf.Primary{MetaIdentifier: "gap free symbol"},
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
																MetaIdentifier: "gap separator",
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
														MetaIdentifier: "gap free symbol",
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
																				MetaIdentifier: "gap separator",
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
						Comments: []string{
							`
The third part of the syntax defines the
removal of bracketed-textual-comments from
gap-free-symbols that form a syntax.
`,
							" see 6.6 ",
						},
						MetaIdentifier: "commentless symbol",
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary:     ebnf.Primary{MetaIdentifier: "terminal character"},
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
															MetaIdentifier: "decimal digit",
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
															MetaIdentifier: "first quote symbol",
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
															MetaIdentifier: "second quote symbol",
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
															MetaIdentifier: "start comment symbol",
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
															MetaIdentifier: "end comment symbol",
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
															MetaIdentifier: "special sequence symbol",
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
															MetaIdentifier: "other character",
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
												MetaIdentifier: "meta identifier",
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
												MetaIdentifier: "terminal string",
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
												MetaIdentifier: "special sequence",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 4.9 "},
						MetaIdentifier: "integer",
						Definitions: ebnf.DefinitionsList{
							{Terms: []ebnf.Term{
								{
									Factor: ebnf.Factor{
										Repetitions: -1,
										Primary:     ebnf.Primary{MetaIdentifier: "decimal digit"},
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
																	MetaIdentifier: "decimal digit",
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
						Comments:       []string{" see 4.14 "},
						MetaIdentifier: "meta identifier",
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
																MetaIdentifier: "meta identifier character",
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
						Comments:       []string{" see 4.15 "},
						MetaIdentifier: "meta identifier character",
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
												MetaIdentifier: "decimal digit",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 4.19 "},
						MetaIdentifier: "special sequence",
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										MetaIdentifier: "special sequence symbol",
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
																MetaIdentifier: "special sequence character",
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
										MetaIdentifier: "special sequence symbol",
									},
								},
							},
						}}},
					},
					{
						Comments:       []string{" see 4.20 "},
						MetaIdentifier: "special sequence character",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "terminal character",
											},
										},
										Exception: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "special sequence symbol",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 6.7 "},
						MetaIdentifier: "comment symbol",
						Definitions: ebnf.DefinitionsList{
							{
								Terms: []ebnf.Term{
									{
										Factor: ebnf.Factor{
											Repetitions: -1,
											Primary: ebnf.Primary{
												MetaIdentifier: "bracketed textual comment",
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
												MetaIdentifier: "other character",
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
												MetaIdentifier: "commentless symbol",
											},
										},
									},
								},
							},
						},
					},
					{
						Comments:       []string{" see 6.8 "},
						MetaIdentifier: "bracketed textual comment",
						Definitions: ebnf.DefinitionsList{{Terms: []ebnf.Term{
							{
								Factor: ebnf.Factor{
									Repetitions: -1,
									Primary: ebnf.Primary{
										MetaIdentifier: "start comment symbol",
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
																MetaIdentifier: "comment symbol",
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
									Primary:     ebnf.Primary{MetaIdentifier: "end comment symbol"},
								},
							},
						}}},
					},
					{
						Comments:       []string{" see 6.9 "},
						MetaIdentifier: "syntax",
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
																MetaIdentifier: "bracketed textual comment",
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
									Primary:     ebnf.Primary{MetaIdentifier: "commentless symbol"},
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
																MetaIdentifier: "bracketed textual comment",
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
														MetaIdentifier: "commentless symbol",
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
																				MetaIdentifier: "bracketed textual comment",
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
