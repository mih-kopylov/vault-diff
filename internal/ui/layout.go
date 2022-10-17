package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/mih-kopylov/vault-diff/internal/vault"
	"github.com/spf13/viper"
)

func mainLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView("main", 0, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf("Vault Diff :: %s", viper.Get("url"))
		v.Wrap = true

	}

	client, err := vault.NewClient()
	if err != nil {
		return err
	}

	secrets, err := vault.GetAvailableSecrets(client)
	if err != nil {
		return err
	}

	v.Clear()
	_, err = fmt.Fprintf(v, `Hello world!
	
	secrets = %v
	
	Counter = %v`, secrets, content.counter)
	if err != nil {
		return err
	}

	return nil
}

func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
