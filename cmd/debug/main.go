package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/query"
	"github.com/midbel/query/cmd/internal/code"
)

func main() {
	tokenize := flag.Bool("t", false, "tokenize")
	flag.Parse()
	if *tokenize {
		s := query.NewScanner(flag.Arg(0))
		for tok := s.Scan(); tok.Type != query.TokEOF; tok = s.Scan() {
			fmt.Println(tok)
		}
		return
	}
	q, err := query.Parse(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code.ExitBadQuery)
	}
	query.Debug(q, os.Stdout)
}
