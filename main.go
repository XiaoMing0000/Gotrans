package main

import (
	fyneapp "fyne.io/fyne/v2/app"

	"gotrans/internal/ui"
)

func main() {
	a := fyneapp.New()
	w := ui.Build(a)
	w.ShowAndRun()
}
