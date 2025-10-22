// Package main is for manual testing of the iso package.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alec-w/ebnf-go/w3c"
)

const sample = "testRule ::= 'one' | 'two' | 'three'"

func main() {
	parser := w3c.New()
	syntax, err := parser.Parse(sample)
	if err != nil {
		//nolint:forbidigo // cmd/cli is for manual testing currently
		fmt.Printf("Error: %s.\n", err)
		os.Exit(1)
	}
	out := new(strings.Builder)
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(syntax); err != nil {
		//nolint:forbidigo // cmd/cli is for manual testing currently
		fmt.Printf("Error: %s.\n", err)
		os.Exit(1)
	}
	//nolint:forbidigo // cmd/cli is for manual testing currently
	fmt.Println(out.String())
}
