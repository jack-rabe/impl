package parser

import (
	"bytes"
	"testing"
)

func TestEmptyInterface(t *testing.T) {
	filename := "file.go"
	var b bytes.Buffer
	b.WriteString(`
type Doer interface {

}
`)
	interfaces, err := getInterfaces(&b, filename)
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

}

func TestInterfaceWithPrivateMethods(t *testing.T) {
	var b bytes.Buffer
	b.WriteString(`
type Doer interface {
	Wag()
	bark(a int)
}
`)
	interfaces, err := getInterfaces(&b, "file.go")
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
	if method != "Wag()" {
		t.Fatalf("Expected method to be Wag(), got %s", method)
	}
}

func TestEmbeddedInterface(t *testing.T) {
	var b bytes.Buffer
	b.WriteString(`
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}
`)
	interfaces, err := getInterfaces(&b, "file.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(interfaces) != 2 {
		t.Fatalf("Expected to find 2 interfaces, got %v\n", len(interfaces))
	}
	exprI := interfaces[1]
	t.Fatal(exprI)
	if len(exprI.Methods) != 2 {
		t.Fatalf("Expected to find 2 public methods on Expr, got %v\n", len(exprI.Methods))
	}
}
