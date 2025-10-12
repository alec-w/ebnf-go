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
	if !assertSlicesEqual(t, expected.Rules, actual.Rules, "rules", "rule", assertRulesEqual) {
		t.Log("Syntax rules were not equal")
		t.Fail()
		return false
	}
	return true
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
		expected.Definitions,
		actual.Definitions,
		"definitions",
		"definition",
		assertDefinitionsEqual,
	) {
		t.Log("Rule definitions were not equal")
		t.Fail()
		failed = true
	}
	return !failed
}

func assertDefinitionsEqual(t *testing.T, expected, actual ebnf.Definition) bool {
	t.Helper()
	if !assertSlicesEqual(t, expected.Terms, actual.Terms, "terms", "term", assertTermsEqual) {
		t.Log("Definition terms were not equal")
		t.Fail()
		return false
	}
	return true
}

func assertTermsEqual(t *testing.T, expected, actual ebnf.Term) bool {
	t.Helper()
	var failed bool
	if !assertFactorsEqual(t, expected.Factor, actual.Factor) {
		t.Log("Term factors were not equal")
		t.Fail()
		failed = true
	}
	if !assertFactorsEqual(t, expected.Exception, actual.Exception) {
		t.Log("Term exceptions were not equal")
		t.Fail()
		failed = true
	}
	return !failed
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
						Definitions: []ebnf.Definition{
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
						Definitions: []ebnf.Definition{
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
						Definitions: []ebnf.Definition{
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
										Primary: ebnf.Primary{RepeatedSequence: []ebnf.Definition{
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
										}},
									},
								},
							}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			syntax, err := ebnf.ParseSyntax(tc.grammar)
			if err != nil {
				t.Fatalf("Got unexpected error %s.", err)
			}
			assertSyntaxesEqual(t, tc.expectedSyntax, syntax)
		})
	}
}
