package main

import (
	"encoding/json"
	"fmt"
	"impl/parser"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const GO_ROOT_DIR = "/opt/homebrew/Cellar/go/1.22.1/libexec/src/"

func main() {
	fmt.Printf("searching %s for interfaces...\n", GO_ROOT_DIR)
	interfaces := make([]parser.GoInterface, 0)
	err := filepath.WalkDir(GO_ROOT_DIR, walkDirFn(&interfaces))
	if err != nil {
		panic(err)
	}
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

func walkDirFn(allInterfaces *[]parser.GoInterface) fs.WalkDirFunc {
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
			interfaces, err := parser.GetInterfaces(path, len(GO_ROOT_DIR))
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
