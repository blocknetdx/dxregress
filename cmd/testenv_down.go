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
	"context"
	"os"
	"os/signal"
	"path"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// testEnvDownCmd attempts to shut down all testenv nodes created in "up".
var testEnvDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Terminate the test environment",
	Long: `This command will attempt to shut down all testenv nodes created in "up"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create docker cli instance
		var err error
		docker, err := client.NewEnvClient()
		if err != nil {
			logrus.Error(err)
			stop()
			return
		}
		defer docker.Close()

		// Path to testenv config dir
		configPath := path.Join(getConfigPath(), "testenv")

		// Create the testenv test environment
		testEnv := chain.NewTestEnv(&chain.EnvConfig{
			ConfigPath:          configPath,
			ContainerPrefix:     testenvPrefix,
			DefaultImage:        "",
			ContainerFilter:     testenvContainerFilter(""),
			ContainerFilterFunc: testenvContainerFilter,
			DockerFileName:      dockerFileName,
		}, docker)

		// Capture container stop success
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		waitChan := make(chan bool, 1)
		go func() {
			// Stop and remove containers
			if err := testEnv.Stop(ctx); err != nil {
				logrus.Error(err)
				waitChan <- false
			} else {
				waitChan <- true
			}
		}()

		// Capture terminal interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		// Block until signals received
		select {
		case <-sigChan:
			logrus.Error("Warning: terminating before resource clean up")
			stop()
			return
		case success := <-waitChan:
			if success {
				logrus.Info("Successfully shutdown testenv")
			} else {
				logrus.Info("Failed to shutdown testenv")
			}
		}
	},
}

func init() {
	testEnvCmd.AddCommand(testEnvDownCmd)
}
