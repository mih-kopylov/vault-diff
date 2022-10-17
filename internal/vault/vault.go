package vault

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

func NewClient() (*api.Client, error) {
	client, err := api.NewClient(&api.Config{Address: viper.GetString("url")})
	client.SetToken(viper.GetString("token"))
	return client, err
}

func GetAvailableSecrets(client *api.Client) ([]string, error) {
	list, err := client.Logical().List(fmt.Sprintf("%v/metadata", viper.GetString("path")))
	if err != nil {
		return nil, err
	}
	var result []string
	for _, element := range list.Data["keys"].([]any) {
		result = append(result, element.(string))
	}
	return result, err
}
