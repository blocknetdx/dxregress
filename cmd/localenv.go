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
	"fmt"

	"github.com/BlocknetDX/dxregress/containers"
	"github.com/spf13/cobra"
)

const localenvPrefix = "dxregress-localenv-"
const localenvContainerImage = "blocknetdx/dxregress:localenv"

var codedir string

// localenvCmd represents the localenv command
var localenvCmd = &cobra.Command{
	Use:   "localenv",
	Short: "Create a test environment from a local codebase",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !containers.IsDockerInstalledAndRunning() {
			stop()
			return
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// init adds the localenv cmd to the root command
func init() {
	RootCmd.AddCommand(localenvCmd)
}

// localEnvContainerFilter returns the regex filter for the localenv containers.
func localEnvContainerFilter(prefix string) string {
	return fmt.Sprintf(`^/%s%s[^\s]+$`, localenvPrefix, prefix)
}
