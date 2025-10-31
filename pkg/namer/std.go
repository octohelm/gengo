//go:generate go run ./__generators__/main.go
package namer

import (
	"bufio"
	"bytes"
)

import (
	_ "embed"
)

//go:embed std.list
var stdPkgList []byte

var std = &defaultImportTracker{
	pathToName: map[string]string{},
	nameToPath: map[string]string{},
}

func init() {
	scanner := bufio.NewScanner(bytes.NewBuffer(stdPkgList))

	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			std.add(line)
		}
	}
}
