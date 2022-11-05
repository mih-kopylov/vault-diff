package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2/json"
)

func NewClient() (*api.Client, error) {
	client, err := api.NewClient(&api.Config{Address: viper.GetString("url")})
	client.SetToken(viper.GetString("token"))
	return client, err
}

func GetSecret(client *api.Client, name string, version int) (string, error) {
	kvSecret, err := client.KVv2(viper.GetString("path")).GetVersion(context.Background(), name, version)
	if err != nil {
		return "", err
	}

	if kvSecret.Data == nil {
		return "", nil
	}

	result, err := json.MarshalIndent(kvSecret.Data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}
