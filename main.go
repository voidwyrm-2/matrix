package main

import (
	_ "embed"
	"image"
	"log"
	"os"
	"strings"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/style"
)

//go:embed version.txt
var version string

var (
	mainWindow    nucular.MasterWindow
	config        matrixConfig
	logFileHandle *os.File
	logger        *log.Logger
)

func main() {
	defer func() {
		if logFileHandle != nil {
			logFileHandle.Close()
		}
	}()

	result, err := initAppResources()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	styleTable, err := result.config.Palette.toTable()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	logger = result.logger
	config = result.config

	mainWindow = nucular.NewMasterWindowSize(0, "Matrix Mod Manager version "+strings.TrimSpace(version), image.Point{750, 600}, updatefn)
	mainWindow.SetStyle(style.FromTable(styleTable, MATRIX_GUI_SCALE))
	mainWindow.Main()
}

var controlAsyncMap = map[string]chan error{}

func updatefn(w *nucular.Window) {
	w.Row(50).Dynamic(1)

	for _, control := range controls {
		if w.Button(label.TA(control.name, "CC"), false) {
			if _, ok := controlAsyncMap[control.name]; !ok {
				ch := make(chan error, 1)

				go func() {
					ch <- control.run(mainWindow, &config)
				}()

				controlAsyncMap[control.name] = ch
			}
		}

		if control.desc != "" {
			w.Label(control.desc, "CC")
		}
	}

	for k, ch := range controlAsyncMap {
		select {
		case err := <-ch:
			handleWindowError(mainWindow, err)
			delete(controlAsyncMap, k)
		default:
		}
	}
}

var controls = []struct {
	name, desc string
	run        func(master nucular.MasterWindow, config *matrixConfig) error
}{
	{
		"Demake",
		"Generates a Matrixfile from a matrix.toml",
		demakeMatrixfile,
	},
	{
		"Make",
		"Generate a matrix.toml from a Matrixfile",
		makeMatrixfile,
	},
	{
		"List",
		"List the mods in a certain matrix.toml",
		listMatrixMods,
	},
	{
		"Sync",
		"Download all mods listed in the matrix.toml",
		syncMatrixMods,
	},
	{
		"Exit",
		"",
		func(master nucular.MasterWindow, config *matrixConfig) error {
			master.Close()
			return nil
		},
	},
}
