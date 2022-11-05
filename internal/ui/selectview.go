package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/spf13/viper"
)

type SelectView struct {
	*gocui.View
	secretIndex int
}

func newSelectView(view *gocui.View) *SelectView {
	result := &SelectView{}
	result.View = view

	kvUrl := fmt.Sprintf("%s/ui/vault/secrets/%s", viper.Get("url"), viper.Get("path"))
	title := fmt.Sprintf("Select secret to compare :: %s", kvUrl)
	result.Title = title
	result.Autoscroll = true
	result.Highlight = true

	return result
}

func (v *SelectView) Draw() {
	v.Clear()
	for _, secret := range content.allSecrets {
		_, _ = fmt.Fprintln(v, secret)
	}
}

func (v *SelectView) SelectNextItem() error {
	x, y := v.Cursor()
	_, maxY := v.Size()
	if y < maxY-1 && v.secretIndex < len(content.allSecrets)-1 {
		y++
		v.secretIndex++
	}
	return v.SetCursor(x, y)
}

func (v *SelectView) SelectPreviousItem() error {
	x, y := v.Cursor()
	if y > 0 && v.secretIndex > 0 {
		y--
		v.secretIndex--
	}
	return v.SetCursor(x, y)
}
