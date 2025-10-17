package ebnf

import (
	"encoding/json"
)

type Syntax struct {
	Rules            []Rule   `json:"rules"`
	TrailingComments []string `json:"trailingComments,omitempty"`
}

type Rule struct {
	Line           int             `json:"line"`
	Comments       []string        `json:"comments,omitempty"`
	MetaIdentifier string          `json:"metaIdentifier"`
	Definitions    DefinitionsList `json:"definitions"`
}

type DefinitionsList = []Definition

type Definition struct {
	Terms []Term `json:"terms"`
}

type Term struct {
	Factor    Factor
	Exception Factor
}

func (t Term) MarshalJSON() ([]byte, error) {
	out := map[string]any{}
	out["factor"] = t.Factor
	if !t.Exception.Primary.IsZero() {
		out["exception"] = t.Exception
	}
	marshalled, err := json.Marshal(out)
	if err != nil {
		return nil, &JsonError{wrapped: err}
	}

	return marshalled, nil
}

type Factor struct {
	Comments    []string
	Repetitions int
	Primary     Primary
}

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
		return nil, &JsonError{wrapped: err}
	}

	return marshalled, nil
}

type Primary struct {
	OptionalSequence DefinitionsList `json:"optionalSequence,omitempty"`
	RepeatedSequence DefinitionsList `json:"repeatedSequence,omitempty"`
	SpecialSequence  string          `json:"specialSequence,omitempty"`
	GroupedSequence  DefinitionsList `json:"groupedSequence,omitempty"`
	MetaIdentifier   string          `json:"metaIdentifier,omitempty"`
	Terminal         string          `json:"terminal,omitempty"`
	Empty            bool            `json:"empty,omitempty"`
}

func (p *Primary) IsZero() bool {
	return p.OptionalSequence == nil && p.RepeatedSequence == nil && p.SpecialSequence == "" &&
		p.GroupedSequence == nil &&
		p.MetaIdentifier == "" &&
		p.Terminal == "" &&
		!p.Empty
}
