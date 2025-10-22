package w3c_test

import (
	"strings"
	"testing"

	"github.com/alec-w/ebnf-go/w3c"
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

func assertSyntaxesEqual(t *testing.T, expected, actual w3c.Syntax) bool {
	t.Helper()
	if assertSlicesEqual(t, expected.Rules, actual.Rules, "rules", "rule", assertRulesEqual) {
		return true
	}
	t.Error("Syntax rules were not equal.")
	return false
}

func assertRulesEqual(t *testing.T, expected, actual w3c.Rule) bool {
	t.Helper()
	var failed bool
	if expected.Symbol != actual.Symbol {
		t.Errorf("Expected rule to have symbol %q but got %q.", expected.Symbol, actual.Symbol)
		failed = true
	}
	return assertExpressionsEqual(t, expected.Expression, actual.Expression) && !failed
}

func assertExpressionsEqual(t *testing.T, expected, actual w3c.Expression) bool {
	t.Helper()
	var failed bool
	if !assertLiteralExpressionsEqual(t, expected.LiteralExpression(), actual.LiteralExpression()) {
		t.Errorf("Literal expressions were not equal")
		failed = true
	}
	if !assertSymbolExpressionsEqual(t, expected.SymbolExpression(), actual.SymbolExpression()) {
		t.Errorf("Symbol expressions were not equal")
		failed = true
	}
	if !assertListExpressionsEqual(t, expected.ListExpression(), actual.ListExpression()) {
		t.Errorf("List expressions were not equal")
		failed = true
	}
	if !assertAlternateExpressionsEqual(t, expected.AlternateExpression(), actual.AlternateExpression()) {
		t.Errorf("Alternate expressions were not equal")
		failed = true
	}
	if !assertExceptionExpressionsEqual(t, expected.ExceptionExpression(), actual.ExceptionExpression()) {
		t.Errorf("Exception expressions were not equal")
		failed = true
	}
	return !failed
}

func assertLiteralExpressionsEqual(t *testing.T, expected, actual *w3c.LiteralExpression) bool {
	t.Helper()
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil {
		t.Error("Got unexpected literal expression.")
		return false
	}
	if actual == nil {
		t.Error("Expected literal expression.")
		return false
	}
	if expected.Literal == actual.Literal {
		return true
	}
	t.Errorf("Expected literal expression to have literal %q but got %q.", expected.Literal, actual.Literal)
	return false
}

func assertSymbolExpressionsEqual(t *testing.T, expected, actual *w3c.SymbolExpression) bool {
	t.Helper()
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil {
		t.Error("Got unexpected symbol expression.")
		return false
	}
	if actual == nil {
		t.Error("Expected symbol expression.")
		return false
	}
	if expected.Symbol == actual.Symbol {
		return true
	}
	t.Errorf("Expected symbol expression to have symbol %q but got %q.", expected.Symbol, actual.Symbol)
	return false
}

func assertListExpressionsEqual(t *testing.T, expected, actual *w3c.ListExpression) bool {
	t.Helper()
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil {
		t.Error("Got unexpected list expression.")
		return false
	}
	if actual == nil {
		t.Error("Expected list expression.")
		return false
	}
	if len(expected.Expressions) != len(actual.Expressions) {
		t.Errorf("Expected %d expressions in list but got %d", len(expected.Expressions), len(actual.Expressions))
		return false
	}
	if assertSlicesEqual(t, expected.Expressions, actual.Expressions, "expressions", "expression", assertExpressionsEqual) {
		return true
	}
	t.Error("List expressions where not equal")
	return false
}

func assertAlternateExpressionsEqual(t *testing.T, expected, actual *w3c.AlternateExpression) bool {
	t.Helper()
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil {
		t.Error("Got unexpected alternate expression.")
		return false
	}
	if actual == nil {
		t.Error("Expected alternate expression.")
		return false
	}
	if len(expected.Expressions) != len(actual.Expressions) {
		t.Errorf("Expected %d expressions in alternates but got %d", len(expected.Expressions), len(actual.Expressions))
		return false
	}
	if assertSlicesEqual(t, expected.Expressions, actual.Expressions, "expressions", "expression", assertExpressionsEqual) {
		return true
	}
	t.Error("Alternate expressions where not equal")
	return false
}

func assertExceptionExpressionsEqual(t *testing.T, expected, actual *w3c.ExceptionExpression) bool {
	t.Helper()
	var failed bool
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil {
		t.Error("Got unexpected exception expression.")
		return false
	}
	if actual == nil {
		t.Error("Expected exception expression.")
		return false
	}
	if !assertExpressionsEqual(t, expected.Match, actual.Match) {
		t.Error("Exception match expression were not equal.")
		failed = true
	}
	if !assertExpressionsEqual(t, expected.Except, actual.Except) {
		t.Error("Exception except expression were not equal.")
		failed = true
	}
	return !failed
}

func TestParserParse(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name           string
		grammar        string
		expectedSyntax w3c.Syntax
	}{
		{
			name:    "simple rule",
			grammar: "testRule ::= 'word'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{Symbol: "testRule", Line: 1, Expression: &w3c.LiteralExpression{Literal: "word"}},
			}},
		},
		{
			name:    "list rule",
			grammar: "testRule ::= 'one' 'two'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.ListExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.LiteralExpression{Literal: "two"},
					}},
				},
			}},
		},
		{
			name:    "alternative rule",
			grammar: "testRule ::= 'one' | 'two'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.LiteralExpression{Literal: "two"},
					}},
				},
			}},
		},
		{
			name:    "mixed list and alternative rule",
			grammar: "testRule ::= 'one' | 'two' 'three'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.ListExpression{Expressions: []w3c.Expression{
							&w3c.LiteralExpression{Literal: "two"},
							&w3c.LiteralExpression{Literal: "three"},
						}},
					}},
				},
			}},
		},
		{
			name:    "flipped mixed list and alternative rule",
			grammar: "testRule ::= 'one' 'two' | 'three'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.ListExpression{Expressions: []w3c.Expression{
							&w3c.LiteralExpression{Literal: "one"},
							&w3c.LiteralExpression{Literal: "two"},
						}},
						&w3c.LiteralExpression{Literal: "three"},
					}},
				},
			}},
		},
		{
			name:    "associativity of list rule",
			grammar: "testRule ::= 'one' 'two' 'three'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.ListExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.LiteralExpression{Literal: "two"},
						&w3c.LiteralExpression{Literal: "three"},
					}},
				},
			}},
		},
		{
			name:    "associativity of alternative rule",
			grammar: "testRule ::= 'one' | 'two' | 'three'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.LiteralExpression{Literal: "two"},
						&w3c.LiteralExpression{Literal: "three"},
					}},
				},
			}},
		},
		{
			name:    "long nested list expressions",
			grammar: "testRule ::= 'one' 'two' 'three' | 'four' 'five' 'six'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.ListExpression{Expressions: []w3c.Expression{
							&w3c.LiteralExpression{Literal: "one"},
							&w3c.LiteralExpression{Literal: "two"},
							&w3c.LiteralExpression{Literal: "three"},
						}},
						&w3c.ListExpression{Expressions: []w3c.Expression{
							&w3c.LiteralExpression{Literal: "four"},
							&w3c.LiteralExpression{Literal: "five"},
							&w3c.LiteralExpression{Literal: "six"},
						}},
					}},
				},
			}},
		},
		{
			name:    "long alternate expressions",
			grammar: "testRule ::= 'one' | 'two' | 'three' 'four' | 'five' | 'six'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.AlternateExpression{Expressions: []w3c.Expression{
						&w3c.LiteralExpression{Literal: "one"},
						&w3c.LiteralExpression{Literal: "two"},
						&w3c.ListExpression{Expressions: []w3c.Expression{
							&w3c.LiteralExpression{Literal: "three"},
							&w3c.LiteralExpression{Literal: "four"},
						}},
						&w3c.LiteralExpression{Literal: "five"},
						&w3c.LiteralExpression{Literal: "six"},
					}},
				},
			}},
		},
		{
			name:    "exception expression",
			grammar: "testRule ::= 'one' - 'two'",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.ExceptionExpression{
						Match:  &w3c.LiteralExpression{Literal: "one"},
						Except: &w3c.LiteralExpression{Literal: "two"},
					},
				},
			}},
		},
		{
			name:    "symbol expression",
			grammar: "testRule ::= AnotherRule",
			expectedSyntax: w3c.Syntax{Rules: []w3c.Rule{
				{
					Symbol: "testRule", Line: 1, Expression: &w3c.SymbolExpression{
						Symbol: "AnotherRule",
					},
				},
			}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parser := w3c.New()
			syntax, err := parser.Parse(tc.grammar)
			if err != nil {
				t.Fatalf("Got unexpected error %s", err)
			}
			assertSyntaxesEqual(t, tc.expectedSyntax, syntax)
		})
	}
}
