package parser

import (
	"bytes"
	"testing"
)

func TestEmptyInterface(t *testing.T) {
	filename := "file.go"
	var b bytes.Buffer
	b.WriteString(`
package do

type Doer interface {
}
`)
	interfaces, err := getInterfaces(b.Bytes(), filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("Expected to find 1 interface, got %v\n", len(interfaces))
	}
	i := interfaces[0]
	if i.Name != "Doer" {
		t.Fatalf("Expected name Doer, got %v\n", i.Name)
	}
	if len(i.Methods) != 0 {
		t.Fatalf("Expected to find 0 methods on Doer, got %v\n", len(i.Methods))
	}
	if i.Filename != filename {
		t.Fatalf("Expected filename to be %s, got %v\n", filename, i.Filename)
	}
	if i.Package != "do" {
		t.Fatalf("Expected Package to be, got %s\n", i.Package)
	}

}

func TestInterfaceWithPrivateMethods(t *testing.T) {
	var b bytes.Buffer
	b.WriteString(`
package do

type Doer interface {
	Wag()
	bark(a int)
}
`)
	interfaces, err := getInterfaces(b.Bytes(), "file.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("Expected to find 1 interface, got %v\n", len(interfaces))
	}
	i := interfaces[0]
	if i.Name != "Doer" {
		t.Fatalf("Expected name Doer, got %v\n", i.Name)
	}
	if len(i.Methods) != 1 {
		t.Fatalf("Expected to find 1 public method on Doer, got %v\n", len(i.Methods))
	}
	method := i.Methods[0]
	if method.Content != "Wag()" {
		t.Fatalf("Expected method to be Wag(), got %s", method)
	}
}

func TestEmbeddedInterface(t *testing.T) {
	var b bytes.Buffer
	b.WriteString(`
package main
type Dog interface {
	Posx() int
}
type Node interface {
	Dog
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
}
`)
	interfaces, err := getInterfaces(b.Bytes(), "file.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(interfaces) != 3 {
		t.Fatalf("Expected to find 3 interfaces, got %v\n", len(interfaces))
	}
	exprI := interfaces[2]
	if len(exprI.Methods) != 3 {
		t.Fatalf("Expected to find 3 public methods on Expr, got %v\n", len(exprI.Methods))
	}
}

func TestPackageWorksWithComments(t *testing.T) {
	filename := "file.go"
	var b bytes.Buffer
	b.WriteString(`
// a comment
package do

type Doer interface {
}
`)
	interfaces, err := getInterfaces(b.Bytes(), filename)
	if err != nil {
		t.Fatal(err)
	}
	i := interfaces[0]
	if i.Package != "do" {
		t.Fatalf("Expected Package to be, got %s\n", i.Package)
	}
}
func TestReturnTypes(t *testing.T) {
	filename := "file.go"
	var b bytes.Buffer
	b.WriteString(`
package dog

type Doer interface {
	DoStuff() int
	Yeet() string
	Jog() bool
	Run()
	Run2() (int, string)
	Run3() func (int, string) bool
}
`)
	interfaces, err := getInterfaces(b.Bytes(), filename)
	expected := []string{"int", "string", "bool", "", "(int, string)", "func (int, string) bool"}
	if err != nil {
		t.Fatal(err)
	}
	i := interfaces[0]
	for idx := range i.Methods {
		returnType := i.Methods[idx].ReturnType
		if returnType != expected[idx] {
			t.Fatalf(`expected return type "%s", got "%s"`, expected[idx], returnType)
		}
	}
}

func TestPackageQualifiersAreAddedToParameters(t *testing.T) {
	filename := "file.go"
	var b bytes.Buffer
	b.WriteString(`
package dog

type Potato int

type Doer interface {
	DoStuff(p Potato, s string)
}
`)
	interfaces, err := getInterfaces(b.Bytes(), filename)
	if err != nil {
		t.Fatal(err)
	}
	m := interfaces[0].Methods[0]
	expected := "DoStuff(p dog.Potato, s string)"
	if m.Content != expected {
		t.Fatalf("expected %s, got %s", expected, m.Content)
	}
}
