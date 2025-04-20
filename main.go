package main

import (
	_ "embed"
	"os"

	"github.com/voidwyrm-2/matrix/cmd"
)

//go:embed version.txt
var version string

func _main() error {
	return cmd.Execute(version)
}

func main() {
	if err := _main(); err != nil {
		// os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
