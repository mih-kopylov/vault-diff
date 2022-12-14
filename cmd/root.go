package cmd

import (
	"context"
	"fmt"
	"github.com/mih-kopylov/vault-diff/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os/signal"
	"syscall"
)

func CreateRootCommand(applicationVersion string) *cobra.Command {
	var result = &cobra.Command{
		Use:     "vault-diff",
		Short:   "Shows unsealed secret changes in diff format like git",
		Version: applicationVersion,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			configureLogrus()
		},
	}

	result.SetVersionTemplate("{{.Version}}")

	// once the APP gets a signal, it will mark the context as Done
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	result.SetContext(ctx)

	result.PersistentFlags().Bool("debug", false, "Enable debug level logging. Hide progress bar as well.")
	utils.BindFlag(result.PersistentFlags().Lookup("debug"), "debug")

	result.PersistentFlags().StringP("url", "u", "", "Address of Vault service.")
	utils.BindFlag(result.PersistentFlags().Lookup("url"), "url")
	utils.MarkFlagRequiredOrFail(result.PersistentFlags(), "url")

	result.PersistentFlags().StringP("token", "t", "", "Token to access Vault service.")
	utils.BindFlag(result.PersistentFlags().Lookup("token"), "token")
	utils.MarkFlagRequiredOrFail(result.PersistentFlags(), "token")

	result.PersistentFlags().StringP("path", "p", "kv", "KV engine path.")
	utils.BindFlag(result.PersistentFlags().Lookup("path"), "path")

	result.AddCommand(CreateDiffCommand())
	result.AddCommand(CreateUiCommand())

	return result
}

func Execute(applicationVersion string) {
	rootCmd := CreateRootCommand(applicationVersion)
	err := rootCmd.Execute()
	if err != nil {
		logrus.Debugf("command failed: %v", err)
	}
}

func init() {
	configureViper()
}

func configureViper() {
	viper.SetConfigName("vault-diff")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("VD")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// ignore case when config file is not found
		} else {
			panic(fmt.Errorf("can't read config: %w", err))
		}
	}
	logrus.WithField("file", viper.ConfigFileUsed()).Debug("config used")
}

func configureLogrus() {
	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugf("debug logging enabled")
	}
}
