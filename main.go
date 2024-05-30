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
	"sync"
)

const GO_ROOT_DIR = "/usr/local/go/src"

var dirToSearch *string

func main() {
	// search user-defined directory if provided or default go installation if not
	dirToSearch = flag.String("path", GO_ROOT_DIR, "the file path to search for interfaces")
	outputFilename := flag.String("output", "interfaces.json", "the file path for output")
	flag.Parse()
	var err error
	*dirToSearch, err = filepath.Abs(*dirToSearch)
	if err != nil {
		panic(err)
	}
	*outputFilename, err = filepath.Abs(*outputFilename)
	if err != nil {
		panic(err)
	}

	fmt.Printf("searching %s for interfaces...\n", *dirToSearch)

	// collect data for all interfaceChan within the directory
	interfaceChan := make(chan parser.GoInterface)
	interfaces := make([]parser.GoInterface, 0)
	var wg sync.WaitGroup
	err = filepath.WalkDir(*dirToSearch, walkDirFn(interfaceChan, &wg))
	if err != nil {
		panic(err)
	}

	go func() {
		wg.Wait()
		close(interfaceChan)
	}()
	for iface := range interfaceChan {
		interfaces = append(interfaces, iface)
	}

	// marshal data and write to disk
	data, err := json.MarshalIndent(interfaces, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(*outputFilename, data, 0666)
	if err != nil {
		panic(err)
	}

	fmt.Printf("successfully wrote data for %d interfaces to %s\n", len(interfaces), *outputFilename)
}

func walkDirFn(allInterfaces chan parser.GoInterface, wg *sync.WaitGroup) fs.WalkDirFunc {
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
		wg.Add(1)
		go func() {
			addInterfacesFromFile(path, allInterfaces)
			wg.Done()
		}()

		return nil
	}
}

func addInterfacesFromFile(path string, allInterfaces chan parser.GoInterface) {
	if strings.Contains(path, ".go") {
		interfaces, err := parser.GetInterfaces(path, len(*dirToSearch)+1)
		if err != nil {
			return
		}

		for _, iface := range interfaces {
			allInterfaces <- iface
		}
	}
}
