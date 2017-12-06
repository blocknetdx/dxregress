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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// testEnvCmd represents the env command
var testEnvCmd = &cobra.Command{
	Use:   "testenv",
	Short: "Create a regression test environment",
	Long: `The regression test environment includes an activator node, servicenode and premined
chain.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("testenv called")

	},
}

func init() {
	RootCmd.AddCommand(testEnvCmd)

	testEnvCmd.PersistentFlags().String("client-version", "-c", "Client version to use for the spun-up nodes.")
}
