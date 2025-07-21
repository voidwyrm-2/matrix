package main

import (
	"cmp"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/rect"
)

type pickFileOptions struct {
	Filter          string
	CaseInsensitive bool
	AllowDotfiles   bool
}

func pickFile(master nucular.MasterWindow, popupFlags nucular.WindowFlags, popupSize rect.Rect, options pickFileOptions) (path string, canceled bool, err error) {
	pathCh := make(chan string, 1)
	errCh := make(chan error, 1)
	canceledCh := make(chan bool, 1)

	matrixExplorerUpdate := createFileExplorerUpdate(pathCh, errCh, canceledCh, options)
	select {
	case e := <-errCh:
		return "", false, e
	default:
	}

	master.PopupOpen("Pick File", popupFlags, popupSize, false, matrixExplorerUpdate)

	return <-pathCh, <-canceledCh, <-errCh
}

func createFileExplorerUpdate(out chan<- string, errout chan<- error, canceled chan<- bool, options pickFileOptions) func(w *nucular.Window) {
	sentErr := true
	sentCanceled := false

	defer func() {
		if !sentErr {
			errout <- nil
		}

		if !sentCanceled {
			canceled <- false
		}
	}()

	filterReg, err := regexp.Compile(options.Filter)
	if err != nil {
		out <- ""
		errout <- err
		sentErr = true
		return nil
	}

	path, err := os.UserHomeDir()
	if err != nil {
		out <- ""
		errout <- err
		sentErr = true
		return nil
	}

	dirHasBeenRead := false
	entries := []os.DirEntry{}

	states := []bool{}

	return func(w *nucular.Window) {
		w.Row(40).Dynamic(1)

		if !dirHasBeenRead {
			innerEntries, err := os.ReadDir(path)
			if err != nil {
				out <- ""
				errout <- err
				sentErr = true
				return
			}

			entries = []os.DirEntry{}

			for _, entry := range innerEntries {
				str := entry.Name()

				if options.CaseInsensitive {
					str = strings.ToLower(str)
				}

				if entry.Type().IsDir() || len(filterReg.FindString(str)) == len(str) {
					if str[0] != '.' || options.AllowDotfiles {
						entries = append(entries, entry)
					}
				}
			}

			slices.SortFunc(entries, func(a, b os.DirEntry) int {
				return cmp.Compare(strings.ToLower(a.Name()), strings.ToLower(b.Name()))
			})

			states = make([]bool, len(entries)+2)

			dirHasBeenRead = true
		}

		if w.SelectableLabel("<Cancel>", "CT", &states[0]) {
			out <- ""
			canceled <- true
			sentCanceled = true
			w.Close()
			return
		}

		if w.SelectableLabel("<Back>", "CT", &states[1]) {
			path = filepath.Dir(path)
			dirHasBeenRead = false
		}

		for i, entry := range entries {
			if w.SelectableLabel(entry.Name(), "CT", &states[i+2]) {
				path = filepath.Join(path, entry.Name())

				if !entry.Type().IsDir() {
					out <- path
					errout <- nil
					w.Close()
					return
				}

				dirHasBeenRead = false
			}
		}
	}
}
