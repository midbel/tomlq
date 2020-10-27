package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/query"
	"github.com/midbel/query/cmd/internal/code"
	"github.com/midbel/toml"
)

func main() {
	kv := flag.Bool("k", false, "print key/value")
	flag.Parse()

	q, err := query.Parse(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, flag.Arg(0), err)
		os.Exit(code.ExitBadQuery)
	}

	doc, err := decodeDocument(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code.ExitBadDoc)
	}
	ifi, err := q.Select(doc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code.ExitBadQuery)
	}
	if len(ifi) == 0 {
		os.Exit(code.ExitEmpty)
	}
	print := nokey
	if *kv {
		print = withkey
	}
	printResults(ifi, print)
}

const (
	jsonExt = ".json"
	tomlExt = ".toml"
)

func decodeDocument(file string) (map[string]interface{}, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var doc = make(map[string]interface{})
	switch ext := filepath.Ext(file); strings.ToLower(ext) {
	case tomlExt:
		err = toml.Decode(r, &doc)
	case jsonExt:
		err = json.NewDecoder(r).Decode(&doc)
	default:
		err = fmt.Errorf("%s: unsupported file type", ext)
	}
	return doc, err
}

func nokey(_ string, value interface{}) {
	fmt.Println(value)
}

func withkey(key string, value interface{}) {
	fmt.Printf("%s = %v\n", key, value)
}

func printResults(rs []query.Result, print func(string, interface{})) {
	for _, r := range rs {
		printResult(strings.Join(r.Paths, "."), r.Value, print)
	}
}

func printResult(key string, value interface{}, print func(string, interface{})) {
	switch ifi := value.(type) {
	case []interface{}:
		for j, i := range ifi {
			printResult(fmt.Sprintf("%s.%d", key, j), i, print)
		}
	case map[string]interface{}:
		for k, v := range ifi {
			printResult(fmt.Sprintf("%s.%s", key, k), v, print)
		}
	default:
		print(key, ifi)
	}
}
