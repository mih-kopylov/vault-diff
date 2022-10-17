package ui

import "github.com/jroimartin/gocui"

func setHotkeys(gui *gocui.Gui) error {
	err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit)
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		content.counter++
		return nil
	})
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		content.counter--
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
