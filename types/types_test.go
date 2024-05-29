package types

import "testing"

func TestCustomTypes(t *testing.T) {
	unqualifiedList := []string{"Doggo", "Cat", "integer", "num", "Harvey"}
	for _, id := range unqualifiedList {
		if IsBuiltIn(id) {
			t.Fatalf("expected %s to be marked as a custom type, but it was not", id)
		}
	}
}

func TestAreAlreadyQualified(t *testing.T) {
	qualifiedList := []string{"int", "string", "rune", "error", "any"}
	for _, id := range qualifiedList {
		if !IsBuiltIn(id) {
			t.Fatalf("expected %s to be marked as a builtin, but it was not", id)
		}
	}
}
