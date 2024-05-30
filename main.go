package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"impl/parser"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const GO_ROOT_DIR = "/usr/local/go/src"

var dirToSearch *string

func main() {
	// search user-defined directory if provided or default go installation if not
	dirToSearch = flag.String("path", GO_ROOT_DIR, "the file path to search for interfaces")
	flag.Parse()
	var err error
	*dirToSearch, err = filepath.Abs(*dirToSearch)
	if err != nil {
		panic(err)
	}

	fmt.Printf("searching %s for interfaces...\n", *dirToSearch)

	// collect data for all interfaces within the directory
	interfaces := make([]parser.GoInterface, 0)
	err = filepath.WalkDir(*dirToSearch, walkDirFn(&interfaces))
	if err != nil {
		panic(err)
	}

	// marshal data and write to disk
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
			panic(err)
		}

		// skip checking certain directories
		excludedWords := []string{"cgo", "internal", "vendor", "_test", "testdata"}
		for _, d := range excludedWords {
			if strings.Contains(path, d) {
				return nil
			}
		}

		// if it is a go file, check for interfaces to add
		if strings.Contains(path, ".go") {
			interfaces, err := parser.GetInterfaces(path, len(*dirToSearch)+1)
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
