package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ebnf "github.com/alec-w/ebnf-go"
)

const sample = `
nonZeroDigit = "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
digit = "0" | nonZeroDigit ;
integer = "0" | nonZeroDigit, { digit } ;
`

func main() {
	parser := ebnf.New()
	syntax, err := parser.Parse(sample)
	if err != nil {
		fmt.Printf("Error: %s.\n", err)
		os.Exit(1)
	}
	out := new(strings.Builder)
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(syntax); err != nil {
		fmt.Printf("Error: %s.\n", err)
		os.Exit(1)
	}
	fmt.Println(out.String())
}
