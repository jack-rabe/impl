package types

import "testing"

func TestCustomTypes(t *testing.T) {
	customList := []string{"*Doggo", "Cat", "integer", "num", "Harvey"}
	for _, id := range customList {
		if IsBuiltIn(id) {
			t.Fatalf("expected %s to be marked as a custom type, but it was not", id)
		}
	}
}

func TestBuiltins(t *testing.T) {
	builtinList := []string{"*int", "string", "rune", "error", "any"}
	for _, id := range builtinList {
		if !IsBuiltIn(id) {
			t.Fatalf("expected %s to be marked as a builtin, but it was not", id)
		}
	}
}
