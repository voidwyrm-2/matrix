package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aarzilli/nucular"
	"github.com/voidwyrm-2/matrix/api/modpack"
)

func demakeMatrixfile(master nucular.MasterWindow, config *matrixConfig) error {
	path, canceled, err := pickFile(master, MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), pickFileOptions{Filter: `matrix\.toml`, CaseInsensitive: true})
	if err != nil {
		return err
	} else if canceled {
		return nil
	}

	pack, err := modpack.FromToml(path, false, false)
	if err != nil {
		return err
	}

	mods := []string{}

	for _, mod := range pack.Mods() {
		entry := ""
		p := mod.ToPublic()

		if p.Slug != "" {
			entry += p.Slug
		} else if p.Id != "" {
			entry += "id " + p.Id
		} else {
			continue
		}

		if p.ForceVersion != "" {
			entry += " " + p.ForceVersion
		}

		mods = append(mods, entry)
	}

	f, err := os.Create(filepath.Join(filepath.Dir(path), "matrixfile"))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n\n", pack.Name(), pack.Version(), pack.GameVersion(), pack.Modloader()) + strings.Join(mods, "\n")))
	return err
}

func makeMatrixfile(master nucular.MasterWindow, config *matrixConfig) error {
	path, canceled, err := pickFile(master, MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), pickFileOptions{Filter: `matrixfile(\.txt)?`, CaseInsensitive: true})
	if err != nil {
		return err
	} else if canceled {
		return nil
	}

	return modpack.FromMatrixfile(path, "matrix.toml")
}

func listMatrixMods(master nucular.MasterWindow, config *matrixConfig) error {
	path, canceled, err := pickFile(master, MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), pickFileOptions{Filter: `matrix.toml`, CaseInsensitive: true})
	if err != nil {
		return err
	} else if canceled {
		return nil
	}

	pack, err := modpack.FromToml(path, false, false)
	if err != nil {
		return err
	}

	upt := func(w *nucular.Window) {
		w.Row(40).Dynamic(1)

		if b := false; w.SelectableLabel("<Close>", "CT", &b) {
			w.Close()
			return
		}

		for _, mod := range pack.Mods() {
			if !mod.IsEmpty() {
				w.Label(fmt.Sprintf("%s ('%s')", mod.Name(), mod.GetIdOrSlug()), "CT")

				if version := mod.ToPublic().Version; version == "" {
					w.Label("version not given", "CT")
				} else {
					w.Label("version "+mod.ToPublic().Version, "CT")
				}
			}
		}
	}

	master.PopupOpen("Matrix Mods", MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), false, upt)

	return nil
}

type syncLogWriter struct {
	guiEntries []string
	logWriter  io.Writer
	mu         sync.Mutex
}

func (slw *syncLogWriter) Write(p []byte) (n int, err error) {
	slw.mu.Lock()

	str := strings.TrimSpace(strings.Split(string(p), MATRIX_LOG_PREFIX)[1])

	slw.guiEntries = append(slw.guiEntries, str)

	slw.mu.Unlock()

	lenB, err := slw.logWriter.Write(p)

	return len(p) + lenB, err
}

func syncMatrixMods(master nucular.MasterWindow, config *matrixConfig) error {
	path, canceled, err := pickFile(master, MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), pickFileOptions{Filter: `matrix.toml`, CaseInsensitive: true})
	if err != nil {
		return err
	} else if canceled {
		return nil
	}

	onlySyncEmpty := true
	ignoreExternals := false

	syncOptions := []ActionConfigOption{
		{
			&onlySyncEmpty,
			"Only sync empty mods",
		},
		{
			&ignoreExternals,
			"Don't attempt to download the external mods",
		},
	}

	canceled = displayActionConfig(master, MATRIX_POPUP_FLAGS, config.DefaultPopupSize.ToRect(), "Sync options", syncOptions)
	if canceled {
		return nil
	}

	pack, err := modpack.FromToml(path, onlySyncEmpty, ignoreExternals)
	if err != nil {
		return err
	}

	syncWriter := &syncLogWriter{
		logWriter: logger.Writer(),
	}

	logger.SetOutput(syncWriter)

	exitCh := make(chan struct{}, 1)

	defer func() {
		exitCh <- struct{}{}
		logger.SetOutput(syncWriter.logWriter)
	}()

	go master.PopupOpen("Sync log", MATRIX_POPUP_FLAGS^nucular.WindowClosable, config.DefaultPopupSize.ToRect(), false, func(w *nucular.Window) {
		select {
		case <-exitCh:
			w.Close()
			return
		default:
		}

		syncWriter.mu.Lock()

		entries := make([]string, len(syncWriter.guiEntries))
		copy(entries, syncWriter.guiEntries)

		syncWriter.mu.Unlock()

		w.Row(15).Dynamic(1)

		for _, entry := range entries {
			w.Label(entry, "LC")
		}
	})

	err = pack.Populate(logger, filepath.Dir(path))
	if err != nil {
		return err
	}

	return pack.ToToml(path)
}
