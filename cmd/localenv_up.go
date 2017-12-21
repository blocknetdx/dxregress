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
	"regexp"
	"strings"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var p_wallets []string

// localenvUpCmd represents the up command
var localenvUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Create a regression test environment from a local codebase",
	Long:  `The path to the codebase must be specified in the command.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the codebase directory
		codedir = args[0]
		if !util.FileExists(codedir) {
			logrus.Errorf("Invalid codebase directory: %s", codedir)
			stop()
			return
		}

		// Create localenv nodes
		var localNodes = []chain.Node{
			{chain.Activator, "activator", chain.NodeContainerName(localenvPrefix, "activator"), "41477", "41427", "41487", chain.GetPortMap("41477", "41476", "41427", "41419", "41487", "41475"), "blocknetdx-cli", false},
			{chain.Sn1, "sn1", chain.NodeContainerName(localenvPrefix, "sn1"), "41478", "41428", "41488", chain.GetPortMap("41478", "41476", "41428", "41419", "41488", "41475"), "blocknetdx-cli", true},
		}
		var xwallets []chain.XWallet
		var xwalletNodes []chain.Node

		// Check that wallets are valid
		if len(p_wallets) == 0 {
			logrus.Warn("No wallets specified. Use the --wallet flag: -w=SYS,address,rpcuser,rpcpassword,wallet-rpc-ipaddress(optional)")
		}
		for _, cmdWallet := range p_wallets {
			wallet, err := xwalletForCmdParameter(cmdWallet)
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
		var err error
		docker, err := client.NewEnvClient()
		if err != nil {
			logrus.Error(err)
			stop()
			return
		}

		// Path to localenv config dir
		configPath := path.Join(getConfigPath(), "localenv")

		// Create the localenv test environment
		testEnv := chain.NewTestEnv(&chain.EnvConfig{
			ConfigPath:          configPath,
			WorkingDirectory:    codedir,
			ContainerPrefix:     localenvPrefix,
			DefaultImage:        localenvContainerImage,
			ContainerFilter:     localEnvContainerFilter(""),
			ContainerFilterFunc: localEnvContainerFilter,
			GenesisPatch:        chain.GenesisPatchV1(),
			DockerFileName:      dockerFileName,
			Activator:           activator,
			Nodes:               localNodes,
			XWallets:            xwallets,
		}, docker)

		// Docker file in the main codebase
		dockerfile := path.Join(codedir, dockerFileName)

		// Support interrupting container build process
		waitChan := make(chan error, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			// Write docker file
			if err := containers.CreateDockerfile(chain.NodeDockerfile(), dockerfile); err != nil {
				waitChan <- err
				return
			}
			// Build image
			logrus.Info("Building localenv container image, please wait...")
			if err := containers.BuildImage(ctx, docker, codedir, path.Base(dockerfile), localenvContainerImage, chain.BlockDefaultWalletDat()); err != nil {
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
	localenvUpCmd.Flags().StringArrayVarP(&p_wallets, "wallet", "w", []string{}, "Specify test wallets: TICKER,address,rpcuser,rpcpassword,rpc-wallet-ipv4address(optional)")
}

// xwalletForCmdParameter returns an XWallet struct from wallet command line parameter.
func xwalletForCmdParameter(cmdWallet string) (chain.XWallet, error) {
	ip := util.GetLocalIP()
	// Remove all spaces from input
	cmdArgs := strings.Split(strings.Replace(cmdWallet, " ", "", -1), ",")
	if len(cmdArgs) < 4 {
		return chain.XWallet{}, errors.New("Incorrect wallet format, the correct format is: TICKER,address,rpcuser,rpcpassword,rpc-wallet-ipv4address(optional)")
	}
	i := 0
	name := cmdArgs[i]; i++
	// TODO User specifiable version
	//version := ""
	//// Assign version if match
	//if ok, _ := regexp.MatchString(`\d+\.\d+\.\d+\.`, cmdArgs[i]); ok {
	//	version = cmdArgs[i]; i++
	//}
	address := cmdArgs[i]; i++
	rpcuser := cmdArgs[i]; i++
	rpcpass := cmdArgs[i]; i++
	// Bring own wallet flag
	bringOwnWallet := false
	if i < len(cmdArgs) {
		if ok, _ := regexp.MatchString(`\d+\.\d+\.\d+\.\d+`, cmdArgs[i]); !ok {
			logrus.Warnf("Wallet %s IPv4 is the wrong format: %s", name, cmdArgs[i])
		} else {
			ip = cmdArgs[i]
			bringOwnWallet = true
		}
	}
	return chain.CreateXWallet(name, "", address, ip, rpcuser, rpcpass, bringOwnWallet), nil
}
