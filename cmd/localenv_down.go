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

	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// localenvDownCmd represents the down command
var localenvDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stops the local test environment",
	Long: `The path to the codebase must be specified in the command.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the codebase directory
		codedir := args[0]
		if !util.FileExists(codedir) {
			logrus.Errorf("Invalid codebase directory: %s", codedir)
			stop()
			return
		}

		// Remove genesis patch to codebase
		if err := util.GitRemovePatch(genesisPatch(), path.Join(getConfigPath(), genesisPatchFile), codedir); err != nil {
			logrus.Error(err)
		}

		// Remove dockerfile
		dockerfile := path.Join(codedir, dockerFilePath)
		if util.FileExists(dockerfile) {
			if err := os.Remove(dockerfile); err != nil {
				logrus.Error(err)
			}
		}

		// Create docker cli instance
		var err error
		docker, err := client.NewEnvClient()
		if err != nil {
			logrus.Error(err)
			stop()
			return
		}
		defer docker.Close()

		// Capture container stop success
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		waitChan := make(chan bool, 1)
		go func() {
			// Stop and remove containers
			if err := stopAllLocalEnvContainers(ctx, docker, false); err != nil {
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
				logrus.Info("Successfully shutdown localenv")
			} else {
				logrus.Info("Failed to shutdown localenv")
			}
		}

	},
}

func init() {
	localenvCmd.AddCommand(localenvDownCmd)
}
