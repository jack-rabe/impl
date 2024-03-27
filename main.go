package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

const GO_ROOT_DIR = "/opt/homebrew/Cellar/go/1.22.1/libexec/src"

func main() {
	interfaces := make([]GoInterface, 0)
	filepath.WalkDir(GO_ROOT_DIR, walkDirFn(&interfaces))
	data, err := json.MarshalIndent(interfaces, "", "  ")
	if err != nil {
		panic(err)
	}
	filename := "interfaces.json"
	err = os.WriteFile(filename, data, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Printf("successfully wrote data for %d interfaces to %s\n", len(interfaces), filename)
}

func walkDirFn(allInterfaces *[]GoInterface) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
		}
		excludedWords := []string{"cgo", "internal", "vendor", "_test", "testdata"}
		for _, d := range excludedWords {
			if strings.Contains(path, d) {
				return nil
			}
		}
		if strings.Contains(path, ".go") {
			interfaces, err := getInterfaces(path)
			if err != nil {
				return nil
			}
			if len(interfaces) > 0 {
				*allInterfaces = append(*allInterfaces, interfaces...)
			}
			return nil
		}
		return nil
	}
}

func getInterfaces(filename string) ([]GoInterface, error) {
	interfaces := make([]GoInterface, 0)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	sourceCode, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(sourceCode) != 0 {
		sourceCode = sourceCode[:len(sourceCode)-1]
	}

	lang := golang.GetLanguage()
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, err
	}

	root := tree.RootNode()
	q := `(
	type_declaration (
		type_spec 
			name: (type_identifier) 
			type: (interface_type)
		)
	)`
	query, err := sitter.NewQuery([]byte(q), lang)
	if err != nil {
		return nil, err
	}
	for i := range root.ChildCount() {
		typeDeclNode := root.Child(int(i))
		queryCursor := sitter.NewQueryCursor()
		queryCursor.Exec(query, typeDeclNode)
		for {
			_, ok := queryCursor.NextMatch()
			if !ok {
				break
			}
			// we found an interface declaration
			goInterface := GoInterface{
				Filename: filename[len(GO_ROOT_DIR)-1:],
				Methods:  []string{},
			}
			typeSpecNode := typeDeclNode.NamedChild(0)

			// get interface name
			idNode := typeSpecNode.NamedChild(0)
			if idNode == nil {
				continue
			}
			interfaceName := idNode.Content(sourceCode)
			if !isUpperCase(interfaceName) {
				continue
			}
			goInterface.Name = interfaceName
			// get interface methods
			typeNode := typeSpecNode.NamedChild(1)
			for j := range typeNode.NamedChildCount() {
				methodNode := typeNode.NamedChild(int(j))
				if methodNode.Type() != "method_spec" {
					continue
				}
				methodName := methodNode.NamedChild(0).Content(sourceCode)
				if isUpperCase(methodName) {
					goInterface.Methods = append(goInterface.Methods, methodNode.Content(sourceCode))
				}
			}
			interfaces = append(interfaces, goInterface)
		}
	}
	return interfaces, nil
}

type GoInterface struct {
	Name     string   `json:"name"`
	Filename string   `json:"filename"`
	Methods  []string `json:"methods"`
}

func (g GoInterface) String() string {
	s := fmt.Sprintf("%s in %s\n", g.Name, g.Filename)
	for _, m := range g.Methods {
		s += fmt.Sprintf("   %v\n", m)
	}
	return s
}

func isUpperCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := s[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}
