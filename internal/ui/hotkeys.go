package ui

import "github.com/jroimartin/gocui"

func setHotkeys(gui *gocui.Gui) error {
	err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, reloadSecrets)
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone, switchTabs)
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, onKeyArrowDown)
	if err != nil {
		return err
	}

	err = gui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, onKeyArrowUp)
	if err != nil {
		return err
	}

	return nil
}

func onKeyArrowDown(_ *gocui.Gui, view *gocui.View) error {
	if view.Name() == SelectLeftView.Name() {
		return SelectLeftView.SelectNextItem()
	}
	if view.Name() == SelectRightView.Name() {
		return SelectRightView.SelectNextItem()
	}
	return nil
}

func onKeyArrowUp(_ *gocui.Gui, view *gocui.View) error {
	if view.Name() == SelectLeftView.Name() {
		return SelectLeftView.SelectPreviousItem()
	}
	if view.Name() == SelectRightView.Name() {
		return SelectRightView.SelectPreviousItem()
	}
	return nil
}

func switchTabs(gui *gocui.Gui, view *gocui.View) error {
	if view.Name() == selectLeftViewName {
		_, err := gui.SetCurrentView(selectRightViewName)
		if err != nil {
			return err
		}
	}

	if view.Name() == selectRightViewName {
		_, err := gui.SetCurrentView(selectLeftViewName)
		if err != nil {
			return err
		}
	}

	return nil
}

func reloadSecrets(_ *gocui.Gui, _ *gocui.View) error {
	err := updateSecrets()
	if err != nil {
		return err
	}
	SelectLeftView.Draw()
	SelectRightView.Draw()

	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
