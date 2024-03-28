package parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// takes a filename and length of the path prefix to ignore when computing
// filepaths and return a slice that contains all public interfaces and their
// public-facing methods
func GetInterfaces(filename string, prefixPathLen int) ([]GoInterface, error) {
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

	truncatedFilename := filename[prefixPathLen:]
	return getInterfaces(sourceCode, truncatedFilename)
}

func getInterfaces(sourceCode []byte, filename string) ([]GoInterface, error) {
	interfaces := make([]GoInterface, 0)
	lang := golang.GetLanguage()
	root, err := getRootNode(lang, sourceCode)

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

	packageName, err := getPackageName(root, sourceCode, lang)
	if err != nil {
		return nil, err
	}

	for i := range root.ChildCount() {
		childNode := root.Child(int(i))

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
					if numChildren == 3 {
						returnNode := methodNode.NamedChild(numChildren - 1)
						// switch returnNode.Type() {
						// case "function_type":
						// 	fmt.Println("func")
						// case "parameter_list":
						// 	fmt.Println("param list")
						// case "type_identifier":
						// 	fmt.Println("type id")
						// case "slice_type":
						// 	fmt.Println("slice")
						// }
						returnType = returnNode.Content(sourceCode)
					}
					methodName := methodNode.NamedChild(0).Content(sourceCode)
					if isUpperCase(methodName) {
						goInterface.Methods = append(goInterface.Methods, method{
							Content:    methodNode.Content(sourceCode),
							ReturnType: returnType,
						})
					}
				// handle inheritance
				case "interface_type_name":
					goInterface.bases = append(goInterface.bases, methodNode.Content(sourceCode))
				}
			}
			interfaces = append(interfaces, goInterface)
		}
	}

	// TODO fix for interfaces that inherit from other files
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
	Name    string `json:"name"`
	Package string `json:"package"`
	// "Superclasses" for this interface
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

func getRootNode(lang *sitter.Language, sourceCode []byte) (*sitter.Node, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, err
	}
	return tree.RootNode(), nil
}

func getPackageName(root *sitter.Node, sourceCode []byte, lang *sitter.Language) (string, error) {
	packageQ := `(package_clause 
		(package_identifier) @id
	)`
	packageQuery, err := sitter.NewQuery([]byte(packageQ), lang)
	if err != nil {
		return "", err
	}

	packageCursor := sitter.NewQueryCursor()
	packageCursor.Exec(packageQuery, root)
	for {
		m, ok := packageCursor.NextMatch()
		if !ok {
			break
		}
		packageName := m.Captures[0].Node
		return packageName.Content(sourceCode), nil
	}
	return "", errors.New("couldn't find a package name")
}
