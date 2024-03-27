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

