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
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// testEnvUpCmd represents the up command
var testEnvUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Creates a test environment with 1 activator and 1 servicenode.",
	Long: `The test environment uses docker to manage the various containers. The activator
is also used as the staker in the environment. The servicenode is connected to the activator via
the "connect=" configuration parameter.

The nodes in this environment have a specific genesis patch applied that allows them to exist in
their own blockchain, separate from both mainnet and testnet.'`,
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the codebase directory
		clientVersion := args[0]
		if match, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, clientVersion); !match {
			logrus.Errorf("Bad version, please use format 1.0.0: %s", clientVersion)
			stop()
			return
		}

		// Create testenv nodes
		var localNodes = chain.DefaultLocalNodes(testenvPrefix)
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
			wc := chain.NodeForWallet(xwallet, testenvPrefix)
			xwalletNodes = append(xwalletNodes, wc)
		}

		activator := chain.NodeForID(chain.Activator, localNodes)

		// Create docker client
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
			DefaultImage:        testenvContainerImage(clientVersion),
			ContainerFilter:     testenvContainerFilter(""),
			ContainerFilterFunc: testenvContainerFilter,
			DockerFileName:      dockerFileName,
			Activator:           activator,
			Nodes:               localNodes,
			XWallets:            xwallets,
		}, docker)

		// Support interrupting container build process
		waitChan := make(chan error, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
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

		logrus.Info("Successfully started testenv")
	},
}

func init() {
	testEnvCmd.AddCommand(testEnvUpCmd)
}
