package main

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

}
