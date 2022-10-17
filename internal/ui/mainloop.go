package ui

import "github.com/jroimartin/gocui"

func MainLoop() error {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	gui.SetManagerFunc(mainLayout)

	err = setHotkeys(gui)
	if err != nil {
		return err
	}

	err = gui.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}
