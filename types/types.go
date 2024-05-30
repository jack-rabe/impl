package types

func IsBuiltIn(t string) bool {
	if t[0] == '*' {
		t = t[1:]
	}

	builtins := []string{
		"string",
		"bool",
		"int",
		"int8",
		"uint8",
		"byte",
		"int16",
		"uint16",
		"int32",
		"rune",
		"uint32",
		"int64",
		"uint64",
		"uint",
		"uintptr",
		"float32",
		"float64",
		"complex64",
		"complex128",
		"error",
		"any",
	}
	for _, builtin := range builtins {
		if builtin == t {
			return true
		}
	}

	return false
}
