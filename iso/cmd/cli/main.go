// Package main is for manual testing of the iso package.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alec-w/ebnf-go/iso"
)

const sample = `
(* A comment *)
(* Another comment *)
(*
A comment across lines
*)
nonZeroDigit = "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
digit = "0" | nonZeroDigit ;
integer = "0" | nonZeroDigit, { digit } ;
`

func main() {
	parser := iso.New()
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
