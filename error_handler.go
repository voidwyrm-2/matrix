package main

import (
	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
)

func handleWindowError(master nucular.MasterWindow, err error) {
	if err != nil {
		didPanic := false

		pan := func() {
			if !didPanic {
				didPanic = true
				panic("A FATAL ERROR HAS OCCURED")
			}
		}

		logger.Println("FATAL ERROR OCCURED:\n" + err.Error())

		// this gets special treatment because it's the error popup
		master.PopupOpen("FATAL ERROR", 0, config.DefaultWindowSize.ToRect(), false, func(w *nucular.Window) {
			w.Master().OnClose(pan)

			w.Row(25).Dynamic(1)

			w.Label("A FATAL ERROR HAS OCCURED", "CT")
			w.Label(err.Error(), "CT")
			w.Label("See log file for further details\nLog file can be found at '"+logFileHandle.Name()+"'", "CT")

			if w.Button(label.T("Ok"), false) {
				w.Close()
				pan()
			}
		})
	}
}
