package vault

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2/json"
	"strings"
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

type SecretKey struct {
	Path string
	Key  string
}

func (s *SecretKey) String() string {
	return fmt.Sprintf("%v%v", s.Path, s.Key)
}

func GetAllSecrets(client *api.Client) ([]SecretKey, error) {
	return getAllSecrets(client, "")
}

func ReadSecretMetadata(client *api.Client, secret string) (*api.KVMetadata, error) {
	metadata, err := client.KVv2(viper.GetString("path")).GetMetadata(context.Background(), secret)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret metadata: secret=%s, %w", secret, err)
	}

	return metadata, nil
}

func getAllSecrets(client *api.Client, path string) ([]SecretKey, error) {
	list, err := client.Logical().List(viper.GetString("path") + "/metadata/" + path)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault secrets: path=%s, %w", path, err)
	}

	keys, exists := list.Data["keys"]
	if !exists {
		return nil, fmt.Errorf("can't find 'keys' key in list response: path=%s", path)
	}

	var result []SecretKey
	for _, value := range keys.([]any) {
		stringKey := value.(string)
		if strings.HasSuffix(stringKey, "/") {
			//it's not a secret but a directory
			subKeys, err := getAllSecrets(client, path+stringKey)
			if err != nil {
				return nil, err
			}

			result = append(result, subKeys...)
		} else {
			//it's a regular secret
			result = append(result, SecretKey{Path: path, Key: stringKey})
		}
	}

	return result, nil
}
