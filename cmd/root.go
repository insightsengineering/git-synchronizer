/*
Copyright 2024 F. Hoffmann-La Roche AG

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.szostok.io/version/extension"
)

var cfgFile string
var logLevel string

type RepositoryPair struct {
	Source      Repository `mapstructure:"source"`
	Destination Repository `mapstructure:"destination"`
}

type Repository struct {
	RepositoryURL string         `mapstructure:"repo"`
	Auth          Authentication `mapstructure:"auth"`
}

type Authentication struct {
	Method    string `mapstructure:"method"`
	TokenName string `mapstructure:"token_name"`
}

// Repository list provided in YAML configuration file.
var inputRepositories []RepositoryPair

var defaultSettings RepositoryPair

var localTempDirectory string

var log = logrus.New()

func setLogLevel() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
	log.SetReportCaller(false)
	customFormatter.FullTimestamp = false
	fmt.Println(`logLevel = "` + logLevel + `"`)
	switch logLevel {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

var rootCmd *cobra.Command

func newRootCommand() {
	rootCmd = &cobra.Command{
		Use:   "git-synchronizer",
		Short: "A tool to synchronize git repositories.",
		Long:  `A tool to synchronize git repositories.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			setLogLevel()

			fmt.Println(`config = "` + cfgFile + `"`)
			inputRepositoriesJSON, err := json.MarshalIndent(inputRepositories, "", "  ")
			checkError(err)
			defaultSettingsJSON, err := json.MarshalIndent(defaultSettings, "", "  ")
			checkError(err)
			log.Trace("inputRepositories = ", string(inputRepositoriesJSON))
			log.Trace("defaultSettings = ", string(defaultSettingsJSON))

			if runtime.GOOS == "windows" {
				localTempDirectory = os.Getenv("TMP") + `\tmp\git-synchronizer`
			} else {
				localTempDirectory = "/tmp/git-synchronizer"
			}

			SetRepositoryAuth(&inputRepositories, defaultSettings)
			ValidateRepositories(inputRepositories)

			err = os.MkdirAll(localTempDirectory, os.ModePerm)
			checkError(err)

			MirrorRepositories(inputRepositories)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.git-synchronizer.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "info",
		"Logging level (trace, debug, info, warn, error). ")

	// Add version command.
	rootCmd.AddCommand(extension.NewVersionCobraCmd())

	cfg := envy.CobraConfig{
		Prefix:     "GITSYNCHRONIZER",
		Persistent: true,
	}
	envy.ParseCobra(rootCmd, cfg)
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in home directory.
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".git-synchronizer")
	}
	// Read in environment variables that match.
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}
}

func Execute() {
	newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initializeConfig() {
	for _, v := range []string{
		"logLevel",
	} {
		// If the flag has not been set in newRootCommand() and it has been set in initConfig().
		// In other words: if it's not been provided in command line, but has been
		// provided in config file.
		// Helpful project where it's explained:
		// https://github.com/carolynvs/stingoftheviper
		if !rootCmd.PersistentFlags().Lookup(v).Changed && viper.IsSet(v) {
			err := rootCmd.PersistentFlags().Set(v, fmt.Sprintf("%v", viper.Get(v)))
			checkError(err)
		}
	}

	// Check if a YAML list of input git repositories has been provided in the configuration file.
	err := viper.UnmarshalKey("repositories", &inputRepositories)
	checkError(err)
	// Read default settings from configuration file.
	err = viper.UnmarshalKey("defaults", &defaultSettings)
	checkError(err)
}
