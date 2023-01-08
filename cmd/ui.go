package cmd

import (
	"github.com/mih-kopylov/vault-diff/internal/ui"
	"github.com/mih-kopylov/vault-diff/vault"
	"github.com/spf13/cobra"
)

func CreateUiCommand() *cobra.Command {

	var result = &cobra.Command{
		Use:   "ui",
		Short: "Runs a UI to observe vault KV storage and to select keys and their versions to compare",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := vault.NewClient()
			if err != nil {
				return err
			}

			err = ui.RunUiApp(client)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return result
}
