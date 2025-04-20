package main

import (
	"os"

	"github.com/voidwyrm-2/matrix/cmd"
)

func _main() error {
	return cmd.Execute()
}

func main() {
	if err := _main(); err != nil {
		// os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
