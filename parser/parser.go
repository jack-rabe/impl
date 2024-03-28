package parser

import (
	"context"
	"fmt"
	"io"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func GetInterfaces(filename string, prefixPathLen int) ([]GoInterface, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	truncatedFilename := filename[prefixPathLen:]
	return getInterfaces(f, truncatedFilename)
}

func getInterfaces(r io.Reader, filename string) ([]GoInterface, error) {
	interfaces := make([]GoInterface, 0)
	sourceCode, err := io.ReadAll(r)
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

	packageQ := `(package_clause (package_identifier))`
	packageQuery, err := sitter.NewQuery([]byte(packageQ), lang)
	if err != nil {
		return nil, err
	}

	var packageName string
	for i := range root.ChildCount() {
		childNode := root.Child(int(i))

		packageCursor := sitter.NewQueryCursor()
		packageCursor.Exec(packageQuery, childNode)
		for {
			_, ok := packageCursor.NextMatch()
			if !ok {
				break
			}
			packageName = childNode.NamedChild(0).Content(sourceCode)
		}

		queryCursor := sitter.NewQueryCursor()
		queryCursor.Exec(query, childNode)
		for {
			_, ok := queryCursor.NextMatch()
			if !ok {
				break
			}
			// we found an interface declaration
			goInterface := GoInterface{
				Package:  packageName,
				Filename: filename,
				Methods:  []method{},
				bases:    []string{},
			}
			typeSpecNode := childNode.NamedChild(0)

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
				switch methodNode.Type() {
				case "method_spec":
					var returnType string
					numChildren := int(methodNode.NamedChildCount())
					m := methodNode.NamedChild(numChildren - 1)
					if numChildren == 3 {
						returnType = m.Content(sourceCode)
					}
					methodName := methodNode.NamedChild(0).Content(sourceCode)
					if isUpperCase(methodName) {
						goInterface.Methods = append(goInterface.Methods, method{
							Content:    methodNode.Content(sourceCode),
							ReturnType: returnType,
						})
					}
				case "interface_type_name":
					goInterface.bases = append(goInterface.bases, methodNode.Content(sourceCode))

				}
			}
			interfaces = append(interfaces, goInterface)
		}
	}
	for idx := range interfaces {
		defineEmbeddedMethods(idx, interfaces)
	}
	return interfaces, nil
}

func defineEmbeddedMethods(idx int, interfaces []GoInterface) []method {
	i := &interfaces[idx]
	if len(i.bases) == 0 {
		return i.Methods
	}
	for _, base := range i.bases {
		for baseIdx, potentialBase := range interfaces {
			if base == potentialBase.Name {
				methods := defineEmbeddedMethods(baseIdx, interfaces)
				i.Methods = append(i.Methods, methods...)
				break
			}
		}
	}
	i.bases = []string{}
	return i.Methods
}

type GoInterface struct {
	Name     string `json:"name"`
	Package  string `json:"package"`
	bases    []string
	Methods  []method `json:"methods"`
	Filename string   `json:"filename"`
}

type method struct {
	Content    string `json:"content"`
	ReturnType string `json:"return_type"`
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
