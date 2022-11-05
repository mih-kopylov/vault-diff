package ui

import (
	"github.com/jroimartin/gocui"
)

func BuildGui() error {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	gui.SetManagerFunc(uiManager)
	gui.Update(func(gui *gocui.Gui) error {
		err := initContent()
		if err != nil {
			return err
		}
		SelectLeftView.Draw()
		SelectRightView.Draw()

		return nil
	})

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

const (
	selectLeftViewName  = "selectLeft"
	selectRightViewName = "selectRight"
)

var (
	SelectLeftView  *SelectView
	SelectRightView *SelectView
)

func uiManager(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	var err error

	v, err := g.SetView(selectLeftViewName, 0, 0, maxX/2-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, err = g.SetCurrentView(selectLeftViewName)
		if err != nil {
			return err
		}
		SelectLeftView = newSelectView(v)
	}

	v, err = g.SetView(selectRightViewName, maxX/2, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		SelectRightView = newSelectView(v)
	}

	return nil
}
