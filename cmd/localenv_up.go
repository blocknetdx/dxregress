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

	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// localenvUpCmd represents the up command
var localenvUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Creates a new test environment from the local codebase",
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

		// Apply genesis patch to codebase
		if err := util.GitApplyPatch(genesisPatch(), path.Join(getConfigPath(), genesisPatchFile), codedir); err != nil {
			logrus.Error(err)
			stop()
			return
		}

		// Write docker file
		dockerfile := path.Join(codedir, dockerFilePath)
		if err := containers.CreateDockerfile(dockerFile(), dockerfile); err != nil {
			logrus.Error(err)
			stop()
			return
		}

		// Build container image
		var err error
		docker, err := client.NewEnvClient()
		if err != nil {
			logrus.Error(err)
			stop()
			return
		}
		defer docker.Close()

		// localenv containers
		type Node struct {
			Name string
			Port string
			RPCPort string
			DebuggerPort string
			Ports nat.PortMap
		}
		localcs := []Node{
			{"activator", "41476", "41426", "41486", getPortMap("41476", "41426", "41486") },
			{"sn1", "41477", "41427", "41487", getPortMap("41477", "41427", "41487") },
			{"sn2", "41478", "41428", "41488", getPortMap("41478", "41428", "41488") },
		}

		// Support interrupting container build process
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		waitChan := make(chan error, 1)
		// Build localenv image, report results on waitChan
		go func() {
			logrus.Info("Building localenv container image, please wait...")
			if err := containers.BuildImage(ctx, docker, codedir, path.Base(dockerfile), containerImage); err != nil {
				waitChan <- err
				return
			}

			// Stop all localenv containers
			logrus.Info("Removing previous localenv containers...")
			if err := stopAllLocalEnvContainers(ctx, docker, true); err != nil {
				logrus.Error(err)
			}

			// Start localenv containers
			for _, c := range localcs {
				if err := containers.CreateAndStart(ctx, docker, containerImage, dxregressContainerName(c.Name), c.Ports); err != nil {
					waitChan <- err
					return
				}
				logrus.Infof("%s node running on %s, rpc on %s, gdb/lldb port on %s", c.Name, c.Port, c.RPCPort, c.DebuggerPort)
			}

			// success
			waitChan <- nil
		}()

		// Capture terminal interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		// Block until signals received
		select {
		case <-sigChan:
			cancel()
		case err := <-waitChan:
			if err != nil {
				logrus.Error(err)
				stop()
				return
			}
		}

		logrus.Info("Sample rpc call: blocknetdx-cli -rpcuser=localenv -rpcpassword=test -rpcport=41426 getinfo")
		logrus.Info("Successfully started localenv")
	},
}

func init() {
	localenvCmd.AddCommand(localenvUpCmd)
}

// getPortMap returns the port map configuration for the specified port.
func getPortMap(port, rpc, debug string) nat.PortMap {
	return nat.PortMap{
		nat.Port("41475/tcp"): []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: debug},
		},
		nat.Port("41476/tcp"): []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: port},
		},
		nat.Port("41419/tcp"): []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: rpc},
		},
	}
}