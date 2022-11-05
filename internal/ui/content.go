package ui

import (
	"github.com/hashicorp/vault/api"
	"github.com/mih-kopylov/vault-diff/internal/vault"
)

var content struct {
	counter               int
	allSecrets            []string
	selectedSecret        string
	selectedSecretContent string
}

var vaultClient *api.Client

func initContent() error {
	client, err := vault.NewClient()
	if err != nil {
		return err
	}
	vaultClient = client

	err = updateSecrets()
	if err != nil {
		return err
	}

	return nil
}

func updateSecrets() error {
	secrets, err := vault.GetAvailableSecrets(vaultClient)
	if err != nil {
		return err
	}
	content.allSecrets = secrets
	content.counter++

	return nil
}
