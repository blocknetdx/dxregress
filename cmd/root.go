// Copyright © 2017 The Blocknet Developers
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var p_config string
var p_version bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dxregress",
	Short: "BlocknetDX regression test tool",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if p_version {
			version()
			return
		}
		cmd.Help()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func init() {
	// Set max procs
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load config
	cobra.OnInitialize(initConfig)

	// Default flags
	RootCmd.PersistentFlags().StringVar(&p_config, "config", "", "config file")
	RootCmd.Flags().BoolVarP(&p_version, "version", "v", false, "Print version")
}

// stop prevents further execution on the command
func stop() {
	os.Exit(1)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	runPath := createConfigPath()

	if p_config != "" {
		// Use config file from the flag.
		viper.SetConfigFile(p_config)
	} else {
		// Search config in current directory
		viper.AddConfigPath(runPath)
		viper.SetConfigName("dxregress")
	}

	viper.SetEnvPrefix("DX")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file:", viper.ConfigFileUsed())
	}

	// Setup logger
	logrus.SetOutput(RootCmd.OutOrStdout())
	lvlHooks := make(logrus.LevelHooks)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp:true})
	if viper.GetBool("DEBUG") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	for _, hooks := range lvlHooks {
		for _, hook := range hooks {
			logrus.AddHook(hook)
		}
	}
}
