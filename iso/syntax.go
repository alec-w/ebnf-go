package iso

import (
	"encoding/json"
)

// Syntax represents a parsed EBNF syntax.
type Syntax struct {
	Rules            []Rule   `json:"rules"`
	TrailingComments []string `json:"trailingComments,omitempty"`
}

// Rule is a single rule in a parsed EBNF syntax.
type Rule struct {
	Line           int             `json:"line"`
	Comments       []string        `json:"comments,omitempty"`
	MetaIdentifier string          `json:"metaIdentifier"`
	Definitions    DefinitionsList `json:"definitions"`
}

// DefinitionsList is an explicit term for a list of definitions.
type DefinitionsList = []Definition

// Definition is a single definition (within a rule or sequence) of a parsed EBNF syntax.
type Definition struct {
	Terms []Term `json:"terms"`
}

// Term is a single term within a definition of a parsed EBNF syntax.
type Term struct {
	Factor    Factor
	Exception Factor
}

// MarshalJSON fulfils the json.Marshaller interface to exclude a Term's Exception if it is empty.
func (t Term) MarshalJSON() ([]byte, error) {
	out := map[string]any{}
	out["factor"] = t.Factor
	if !t.Exception.Primary.IsZero() {
		out["exception"] = t.Exception
	}
	marshalled, err := json.Marshal(out)
	if err != nil {
		return nil, &JSONError{wrapped: err}
	}

	return marshalled, nil
}

// Factor is the primary part of an EBNF term or its exception.
type Factor struct {
	Comments    []string
	Repetitions int
	Primary     Primary
}

// MarshalJSON fulfils the json.Marshaller interface to exclude a Factor's Repetitions field if there are no
// repetitions.
func (f Factor) MarshalJSON() ([]byte, error) {
	out := map[string]any{}
	if len(f.Comments) > 0 {
		out["comments"] = f.Comments
	}
	out["primary"] = f.Primary
	if f.Repetitions >= 0 {
		out["repetitions"] = f.Repetitions
	}
	marshalled, err := json.Marshal(out)
	if err != nil {
		return nil, &JSONError{wrapped: err}
	}

	return marshalled, nil
}

// Primary is the core part of an EBNF syntax term, which will represent one of
// - an optional sequence (0 or 1 instances of a sequence of definitions)
// - a repeated sequence (0 or more repetitions of a sequence of definitions)
// - a special sequence (a sequence of characters described in a form outside the scope of EBNF)
// - a grouped sequence (an instance of a sequence of definitions)
// - a meta identifier (a reference to another rule of the syntax)
// - a terminal (a string of characters)
// - empty (an empty sequence of definitions).
type Primary struct {
	OptionalSequence DefinitionsList `json:"optionalSequence,omitempty"`
	RepeatedSequence DefinitionsList `json:"repeatedSequence,omitempty"`
	SpecialSequence  string          `json:"specialSequence,omitempty"`
	GroupedSequence  DefinitionsList `json:"groupedSequence,omitempty"`
	MetaIdentifier   string          `json:"metaIdentifier,omitempty"`
	Terminal         string          `json:"terminal,omitempty"`
	Empty            bool            `json:"empty,omitempty"`
}

// IsZero returns false if all the fields within the Primary are their empty values.
func (p *Primary) IsZero() bool {
	return p.OptionalSequence == nil && p.RepeatedSequence == nil && p.SpecialSequence == "" &&
		p.GroupedSequence == nil &&
		p.MetaIdentifier == "" &&
		p.Terminal == "" &&
		!p.Empty
}
