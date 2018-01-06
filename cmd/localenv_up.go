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
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// localenvUpCmd represents the up command
var localenvUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Create a regression test environment from a local codebase",
	Long:  `An environment with an activator, servicenode, and two BLOCK traders will
be setup for regression testing. The path to the codebase must be specified in the command.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		// Obtain the codebase directory
		localenvCodeDir, err := filepath.Abs(args[0])
		if err != nil {
			logrus.Errorf("Invalid codebase directory: %s", localenvCodeDir)
			stop()
			return
		} else if _, err := os.Stat(localenvCodeDir); os.IsNotExist(err) {
			logrus.Errorf("Invalid codebase directory: %s", localenvCodeDir)
			stop()
			return
		}

		// Create localenv nodes
		var localNodes = chain.DefaultLocalNodes(localenvPrefix)
		var xwallets []chain.XWallet
		var xwalletNodes []chain.Node

		// Setup trader wallets
		for _, node := range localNodes {
			if node.ID != chain.Trader {
				continue
			}
			wallet, err := chain.XWalletForCmdParameter(fmt.Sprintf("%s,%s,%s,%s", "BLOCK", node.Address, "localenv", "test"))
			if err != nil || !chain.SupportsWallet(wallet.Name) {
				logrus.Errorf("Unsupported wallet %s", wallet.Name)
				stop()
				return
			}
			wallet.Port = node.Port
			wallet.RPCPort = node.RPCPort
			wallet.BringOwn = true
			xwallets = append(xwallets, wallet)
		}

		// Check that wallets are valid
		if len(p_wallets) == 0 {
			logrus.Warn("No wallets specified. Use the --wallet flag: -w=SYS,address,rpcuser,rpcpassword,wallet-rpc-ipaddress(optional)")
		}
		// Setup xwallets
		for _, cmdWallet := range p_wallets {
			wallet, err := chain.XWalletForCmdParameter(cmdWallet)
			if err != nil || !chain.SupportsWallet(wallet.Name) {
				logrus.Errorf("Unsupported wallet %s", wallet.Name)
				stop()
				return
			}
			xwallets = append(xwallets, wallet)
		}

		// Create xwallet nodes
		for _, xwallet := range xwallets {
			// Ignore BYOW nodes (bring your own wallet)
			if xwallet.BringOwn {
				continue
			}
			// Create node from xwallet
			wc := chain.NodeForWallet(xwallet, localenvPrefix)
			xwalletNodes = append(xwalletNodes, wc)
		}

		activator := chain.NodeForID(chain.Activator, localNodes)

		// Build container image
		docker, err := client.NewEnvClient()
		if err != nil {
			logrus.Error(err)
			stop()
			return
		}
		defer docker.Close()

		// Path to localenv config dir
		configPath := path.Join(getConfigPath(), "localenv")

		// Create the localenv test environment
		testEnv := chain.NewTestEnv(&chain.EnvConfig{
			ConfigPath:          configPath,
			ContainerPrefix:     localenvPrefix,
			DefaultImage:        localenvContainerImage,
			ContainerFilter:     localenvContainerFilter(""),
			ContainerFilterFunc: localenvContainerFilter,
			DockerFileName:      dockerFileName,
			Activator:           activator,
			Nodes:               localNodes,
			XWallets:            xwallets,
		}, docker)

		// Docker file in the main codebase
		dockerfile := path.Join(localenvCodeDir, dockerFileName)

		// Support interrupting container build process
		waitChan := make(chan error, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			// Copy wallet to config path
			if err := ioutil.WriteFile(path.Join(configPath, "wallet.dat"), chain.BlockDefaultWalletDat(), 0644); err != nil {
				waitChan <- err
				return
			}
			// Apply genesis patch to codebase
			if err := util.GitApplyPatch(chain.GenesisPatchV1(), path.Join(getConfigPath(), chain.GenesisPatchFile), localenvCodeDir); err != nil {
				waitChan <- err
				return
			}
			// Write docker file
			if err := containers.CreateDockerfile(chain.NodeDockerfile(), dockerfile); err != nil {
				waitChan <- err
				return
			}
			// Build image
			logrus.Info("Building localenv container image, please wait...")
			if err := containers.BuildImage(ctx, docker, localenvCodeDir, path.Base(dockerfile), localenvContainerImage, chain.BlockDefaultWalletDat()); err != nil {
				waitChan <- err
				return
			}

			// Start the test environment
			if err := testEnv.Start(ctx); err != nil {
				waitChan <- err
				return
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

		// Print sample rpc calls
		allContainers := append(localNodes, xwalletNodes...)
		for _, node := range allContainers {
			logrus.Infof("Sample rpc call %s: docker exec %s %s getinfo", node.ShortName, node.Name, node.CLI)
		}

		// Print location of test blocknet conf
		logrus.Infof("Test blocknetdx.conf file here: %s", chain.TestBlocknetConfFile(configPath))
		if len(xwallets) > 0 {
			logrus.Infof("Wallets enabled: %s", strings.Join(chain.XWalletList(xwallets), ","))
		}

		logrus.Info("Successfully started localenv")
	},
}

func init() {
	localenvCmd.AddCommand(localenvUpCmd)
}
