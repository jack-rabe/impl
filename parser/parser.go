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
	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(src) != 0 {
		src = src[:len(src)-1]
	}

	truncatedFilename := filename[prefixPathLen:]
	return getInterfaces(src, truncatedFilename)
}

func getInterfaces(src []byte, filename string) ([]GoInterface, error) {
	interfaces := make([]GoInterface, 0)
	lang := golang.GetLanguage()
	root, err := getRootNode(lang, src)

	packageName, err := getPackageName(root, src, lang)
	if err != nil {
		return nil, err
	}

	q := `( type_declaration
		( type_spec 
			name: (type_identifier) @name
			type: (interface_type) @type
		)
	)`
	query, err := sitter.NewQuery([]byte(q), lang)
	if err != nil {
		return nil, err
	}
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query, root)
	for {
		m, ok := queryCursor.NextMatch()
		if !ok {
			break
		}

		interfaceName := m.Captures[0].Node.Content(src)
		if !isUpperCase(interfaceName) {
			continue
		}
		interfaceTypeN := m.Captures[1].Node
		methods, err := getMethods(interfaceTypeN, src, lang)
		if err != nil {
			panic(err)
		}

		interfaces = append(interfaces, GoInterface{
			Name:     interfaceName,
			Package:  packageName,
			Filename: filename,
			Methods:  methods,
			bases:    []string{},
		})
	}

	// TODO fix for interfaces that inherit from other files
	for idx := range interfaces {
		defineEmbeddedMethods(idx, interfaces)
	}
	return interfaces, nil
}

// returns the a list of the methods on an interface, given an inteface_type_node
func getMethods(interfaceTypeNode *sitter.Node, src []byte, lang *sitter.Language) ([]method, error) {
	methods := []method{}

	methodQ := `( method_spec
			name: (field_identifier) @name
			parameters: (parameter_list) @params
			result: (_)? @return
		) @content
		`
	methodQuery, err := sitter.NewQuery([]byte(methodQ), lang)
	if err != nil {
		return nil, err
	}
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(methodQuery, interfaceTypeNode)
	for {
		methodMatch, ok := queryCursor.NextMatch()
		if !ok {
			break
		}

		methodName := methodMatch.Captures[1].Node.Content(src)
		if !isUpperCase(methodName) {
			continue
		}
		interfaceMethod := method{
			Content: methodMatch.Captures[0].Node.Content(src),
			Name:    methodName,
			Params:  []string{},
		}

		parametersN := methodMatch.Captures[2].Node
		interfaceMethod.Params = append(interfaceMethod.Params, parametersN.Content(src))
		hasReturnType := len(methodMatch.Captures) == 4
		if hasReturnType {
			interfaceMethod.ReturnType = methodMatch.Captures[3].Node.Content(src)
		}
		methods = append(methods, interfaceMethod)
	}

	return methods, nil
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
	Content    string   `json:"content"`
	Name       string   `json:"name"`
	Params     []string `json:"params"`
	ReturnType string   `json:"return_type"`
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

func getRootNode(lang *sitter.Language, src []byte) (*sitter.Node, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, err
	}
	return tree.RootNode(), nil
}

func getPackageName(root *sitter.Node, src []byte, lang *sitter.Language) (string, error) {
	packageQ := `(package_clause 
		(package_identifier) @id
	)`
	packageQuery, err := sitter.NewQuery([]byte(packageQ), lang)
	if err != nil {
		return "", err
	}

	packageCursor := sitter.NewQueryCursor()
	packageCursor.Exec(packageQuery, root)
	m, ok := packageCursor.NextMatch()
	if !ok {
		return "", errors.New("couldn't find a package name")
	}
	packageName := m.Captures[0].Node
	return packageName.Content(src), nil
}
