package main

import (
	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/rect"
)

type ActionConfigOption struct {
	State *bool
	Desc  string
}

func displayActionConfig(master nucular.MasterWindow, popupFlags nucular.WindowFlags, popupSize rect.Rect, popupName string, options []ActionConfigOption) (canceled bool) {
	canceledCh := make(chan bool, 1)

	configUpdate := createActionConfigUpdate(options, canceledCh)

	master.PopupOpen(popupName, popupFlags^nucular.WindowClosable, popupSize, false, configUpdate)

	return <-canceledCh
}

func createActionConfigUpdate(options []ActionConfigOption, canceled chan<- bool) func(w *nucular.Window) {
	cancelButtonState := false
	okButtonState := false

	return func(w *nucular.Window) {
		w.Row(20).Dynamic(1)

		if w.SelectableLabel("<Cancel>", "CT", &cancelButtonState) {
			canceled <- true
			w.Close()
			return
		}

		if w.SelectableLabel("<Ok>", "CT", &okButtonState) {
			canceled <- false
			w.Close()
			return
		}

		for _, opt := range options {
			w.SelectableLabel(opt.Desc, "CT", opt.State)
		}
	}
}
