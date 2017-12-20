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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cliPath = "src/blocknetdx-cli"

var p_wallets []string

var localWallets []chain.XWallet

// localenvUpCmd represents the up command
var localenvUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Creates a new test environment from the local codebase",
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
			localWallets = append(localWallets, wallet)
		}

		// Check if cli exists
		cliFilePath := path.Join(codedir, cliPath)
		if !util.FileExists(cliFilePath) {
			logrus.Errorf("blocknetdx-cli missing from %s did you build first?", cliFilePath)
			stop()
			return
		}

		// Apply genesis patch to codebase
		if err := util.GitApplyPatch(chain.GenesisPatchV1(), path.Join(getConfigPath(), genesisPatchFile), codedir); err != nil {
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

		// Write test blocknetdx.conf file
		testLocalenvDir := path.Dir(testBlocknetConfFile())
		if err := os.MkdirAll(testLocalenvDir, 0775); err != nil {
			logrus.Error(errors.Wrapf(err, "Failed to create localenv directory %s", testLocalenvDir))
			stop()
			return
		}
		if err := ioutil.WriteFile(testBlocknetConfFile(), []byte(testBlocknetConf()), 0644); err != nil {
			logrus.Error(errors.Wrapf(err, "Failed to write localenv blocknetdx.conf %s", testBlocknetConfFile()))
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

		// Support interrupting container build process
		waitChan := make(chan error, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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
			for _, c := range localContainers {
				if err := containers.CreateAndStart(ctx, docker, containerImage, c.Name, c.Ports); err != nil {
					waitChan <- err
					return
				}
				if c.DebuggerPort != "" {
					logrus.Infof("%s node running on %s, rpc on %s, gdb/lldb port on %s", c.Name, c.Port, c.RPCPort, c.DebuggerPort)
				} else {
					logrus.Infof("%s node running on %s, rpc on %s", c.Name, c.Port, c.RPCPort)
				}
			}

			// Start wallet containers
			for _, w := range localWallets {
				// Ignore BYOW nodes (bring your own wallet)
				if w.BringOwn {
					continue
				}
				// Create node from xwallet
				wc := walletNode(w)
				xwalletContainers = append(xwalletContainers, wc)
				if err := containers.CreateAndStart(ctx, docker, w.Container, wc.Name, wc.Ports); err != nil {
					waitChan <- err
					return
				}
				if wc.DebuggerPort != "" {
					logrus.Infof("%s node running on %s, rpc on %s, gdb/lldb port on %s", wc.Name, wc.Port, wc.RPCPort, wc.DebuggerPort)
				} else {
					logrus.Infof("%s node running on %s, rpc on %s", wc.Name, wc.Port, wc.RPCPort)
				}
			}

			logrus.Info("Waiting for localenv to be ready...")
			if err := waitForLoadenv(ctx, localContainers); err != nil {
				waitChan <- err
				return
			}

			// Setup blockchain
			if err := setupChain(ctx, docker); err != nil {
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

		allContainers := append(localContainers, xwalletContainers...)
		for _, node := range allContainers {
			logrus.Infof("Sample rpc call %s: docker exec %s %s getinfo", node.ShortName, node.Name, node.CLI)
		}
		logrus.Infof("Test blocknetdx.conf file here: %s", testBlocknetConfFile())
		if len(localWallets) > 0 {
			logrus.Infof("Wallets enabled: %s", strings.Join(xbridgeWalletList(localWallets), ","))
		}
		logrus.Info("Successfully started localenv")
	},
}

func init() {
	localenvCmd.AddCommand(localenvUpCmd)
	localenvUpCmd.Flags().StringArrayVarP(&p_wallets, "wallet", "w", []string{}, "Specify test wallets: TICKER,address,rpcuser,rpcpassword,rpc-wallet-ipv4address(optional)")
}

// waitForLoadenv will block for a maximum of 30 seconds until the local environment is ready. The
// getinfo rpc call is checked once every 2 seconds. This method returns if getinfo returns no
// error.
func waitForLoadenv(parentContext context.Context, nodes []Node) error {
	// Wait max 30 seconds for environment to provision
	ctx, cancel := context.WithTimeout(parentContext, time.Second*45)
	defer cancel()

	waitChan := make(chan error, 1)
	waitChanCancel := make(chan bool, 1)
	go func() {
	Done:
		for {
			select {
			case <-waitChanCancel:
				break Done
			default:
				ready := true
				for _, node := range nodes {
					cmd := rpcCommand(node.Name, node.CLI, "getinfo")
					if err := cmd.Run(); err != nil {
						if viper.GetBool("DEBUG") {
							logrus.Debug(err)
						}
						ready = false
					}
				}
				if ready {
					waitChan <- nil
					break Done
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()

	select {
	case err := <-waitChan:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		waitChanCancel <- true
		if ctx.Err() != nil {
			return errors.Wrap(ctx.Err(), "Timeout occurred while waiting for localenv to start up")
		}
	}

	return nil
}

// setupChain will setup the DX environment, copy all configuration files, test RPC connectivity.
func setupChain(ctx context.Context, docker *client.Client) error {
	// activator wallet address: y5zBd8oLQSnTjChTUCfRieTAp5Z31bRwEV key: cQiWHyehhhsRFYadBpj5wQRU9HU23GtHSjyPY2hBLccHWeNq6iTY
	// sn1 alias address: y3DT9bZ69AjvdQFzYTCSpFgT9wJcRpHi7T key: cRdLcWroNyJPJ1BH4Q24pamDQtE3JNdm7tGQoD6mm9brqpYuX1dC
	// sn2 alias address: yF2E6wPBc1YosrGUMhgoet5zPat1A4Z87d key: cMn9aiQGBYqeRzRuTFAModv459UQNxGsXkgPSRQ1W7XwGdGCp1JB

	// Nodes
	activator := localContainerForNode(Activator)
	sn1 := localContainerForNode(Sn1)
	sn2 := localContainerForNode(Sn2)

	activatorC := containers.FindContainer(docker, activator.Name)
	sn1C := containers.FindContainer(docker, sn1.Name)
	sn2C := containers.FindContainer(docker, sn2.Name)

	// First import test address into alias and then generate test coin
	cmd := blockRPCCommands(activator.Name, []string{"importprivkey cQiWHyehhhsRFYadBpj5wQRU9HU23GtHSjyPY2hBLccHWeNq6iTY coin", "setgenerate true 25"})
	if output, err := cmd.Output(); err != nil || string(output) == "" {
		return errors.Wrap(err, "Failed to generate first 25 blocks")
	} else {
		logrus.Debug(string(output))
	}

	// Import alias addresses
	cmd2 := blockRPCCommands(activator.Name, []string{"importprivkey cRdLcWroNyJPJ1BH4Q24pamDQtE3JNdm7tGQoD6mm9brqpYuX1dC sn1", "importprivkey cMn9aiQGBYqeRzRuTFAModv459UQNxGsXkgPSRQ1W7XwGdGCp1JB sn2"})
	if output, err := cmd2.Output(); err != nil {
		return errors.Wrap(err, "Failed to import alias addresses")
	} else {
		logrus.Debug(string(output))
	}

	// Send 5k servicenode coin to each alias
	cmd3 := blockRPCCommands(activator.Name, []string{"sendtoaddress y3DT9bZ69AjvdQFzYTCSpFgT9wJcRpHi7T 5000", "sendtoaddress yF2E6wPBc1YosrGUMhgoet5zPat1A4Z87d 5000"})
	if output, err := cmd3.Output(); err != nil {
		return errors.Wrap(err, "Failed to send 5k servicenode coin")
	} else {
		logrus.Debug(string(output))
	}

	// Break up 10K inputs into 2.5k inputs to help with staking
	cmd4S := make([]string, 75)
	for i := 0; i < len(cmd4S); i++ {
		cmd4S[i] = "sendtoaddress y5zBd8oLQSnTjChTUCfRieTAp5Z31bRwEV 2500"
	}
	cmd4 := blockRPCCommands(activator.Name, cmd4S)
	if output, err := cmd4.Output(); err != nil {
		return errors.Wrap(err, "Failed to split coin")
	} else {
		logrus.Debug(string(output))
	}

	// Generate last 25 PoW blocks
	cmd5 := blockRPCCommand(activator.Name, "setgenerate true 25")
	if output, err := cmd5.Output(); err != nil {
		return errors.Wrap(err, "Failed to generate blocks 25-50")
	} else {
		logrus.Debug(string(output))
	}

	// Obtain servicenode keys
	var keys []string
	cmd6A := blockRPCCommand(sn1.Name, "servicenode genkey")
	if output, err := cmd6A.Output(); err != nil {
		return errors.Wrap(err, "Failed to call genkey on sn1")
	} else {
		keys = append(keys, strings.TrimSpace(string(output)))
	}
	cmd6B := blockRPCCommand(sn2.Name, "servicenode genkey")
	if output, err := cmd6B.Output(); err != nil {
		return errors.Wrap(err, "Failed to call genkey on sn2")
	} else {
		keys = append(keys, strings.TrimSpace(string(output)))
	}

	// Setup activator servicenode.conf
	type OutputsResponse struct {
		TxID  string `json:"txhash"`
		TxPos int    `json:"outputidx"`
	}
	cmd7 := blockRPCCommand(activator.Name, "servicenode outputs")
	output, err := cmd7.Output()
	if err != nil {
		return errors.Wrap(err, "Failed to parse servicenode outputs")
	}
	var outputs []OutputsResponse
	if err := json.Unmarshal(output, &outputs); err != nil {
		return errors.Wrap(err, "Failed to parse servicenode outputs")
	}

	// Servicenodes
	ssn1 := SNode{ID: Sn1, Alias: sn1.ShortName, IP: sn1.IP(), Key: keys[0], CollateralID: outputs[0].TxID, CollateralPos: strconv.Itoa(outputs[0].TxPos)}
	ssn2 := SNode{ID: Sn2, Alias: sn2.ShortName, IP: sn2.IP(), Key: keys[1], CollateralID: outputs[1].TxID, CollateralPos: strconv.Itoa(outputs[1].TxPos)}
	snodes := []SNode{ssn1, ssn2}

	// Max wait time for all commands below
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Generate activator servicenode.conf
	snConf := servicenodeConf(snodes)
	// Copy activator servicenode.conf
	if servicenodeConf, err := util.CreateTar(map[string][]byte{"servicenode.conf": []byte(snConf)}); err == nil {
		if err := docker.CopyToContainer(ctx, activatorC.ID, "/opt/blockchain/dxregress/testnet4/", servicenodeConf, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write servicenode.conf to activator")
		}
	} else {
		return errors.Wrap(err, "Failed to write servicenode.conf to activator")
	}

	// Update activator blocknetdx.conf
	blocknetConfActivator := blocknetdxConf(Activator, localContainers, "")
	if bufActivator, err := util.CreateTar(map[string][]byte{"blocknetdx.conf": []byte(blocknetConfActivator)}); err == nil {
		if err := docker.CopyToContainer(ctx, activatorC.ID, "/opt/blockchain/config/", bufActivator, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write blocknetdx.conf to activator")
		}
	} else {
		return errors.Wrap(err, "Failed to write blocknetdx.conf to activator")
	}

	// Update sn1 blocknetdx.conf
	blocknetConfSn1 := blocknetdxConf(Sn1, localContainers, ssn1.Key)
	if bufSn1, err := util.CreateTar(map[string][]byte{"blocknetdx.conf": []byte(blocknetConfSn1)}); err == nil {
		if err := docker.CopyToContainer(ctx, sn1C.ID, "/opt/blockchain/config/", bufSn1, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write blocknetdx.conf to sn1")
		}
	} else {
		return errors.Wrap(err, "Failed to write blocknetdx.conf to sn1")
	}

	// Update sn2 blocknetdx.conf
	blocknetConfSn2 := blocknetdxConf(Sn2, localContainers, ssn2.Key)
	if bufSn2, err := util.CreateTar(map[string][]byte{"blocknetdx.conf": []byte(blocknetConfSn2)}); err == nil {
		if err := docker.CopyToContainer(ctx, sn2C.ID, "/opt/blockchain/config/", bufSn2, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write blocknetdx.conf to sn2")
		}
	} else {
		return errors.Wrap(err, "Failed to write blocknetdx.conf to sn2")
	}

	// Write sn1 & sn2 xbridge.conf
	xbridgeConfSnode := xbridgeConf(localWallets)
	if bufSn, err := util.CreateTar(map[string][]byte{"xbridge.conf": []byte(xbridgeConfSnode)}); err == nil {
		if err := docker.CopyToContainer(ctx, sn1C.ID, "/opt/blockchain/dxregress/", bufSn, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write xbridge.conf to sn1")
		}
		if err := docker.CopyToContainer(ctx, sn2C.ID, "/opt/blockchain/dxregress/", bufSn, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write xbridge.conf to sn2")
		}
	} else {
		return errors.Wrap(err, "Failed to write xbridge.conf")
	}

	// Stop activator
	if err := containers.StopContainer(ctx, docker, activatorC.ID); err != nil {
		return err
	}

	// Restart all nodes except for wallet nodes
	if err := containers.RestartContainers(ctx, docker, localEnvContainerFilter("sn")); err != nil {
		return err
	}
	if err := containers.RestartContainers(ctx, docker, localEnvContainerFilter("act")); err != nil {
		return err
	}

	// Wait for service nodes to be ready
	logrus.Info("Waiting for nodes to be ready...")
	allContainers := append(localContainers, xwalletContainers...)
	if err := waitForLoadenv(ctx, allContainers); err != nil {
		return err
	}

	// Start servicenodes
	if err := startAllServicenodes(activator.Name); err != nil {
		return err
	}

	// Wait before restarting staker
	time.Sleep(10 * time.Second)

	// Restart the activator to trigger staking
	if err := containers.RestartContainers(ctx, docker, localEnvContainerFilter("act")); err != nil {
		return err
	}

	// Wait for activator to be ready
	logrus.Info("Waiting for activator to be ready...")
	if err := waitForLoadenv(ctx, []Node{activator}); err != nil {
		return err
	}

	// Call start servicenodes a second time to make sure they're started
	// Wait before re-running snode command
	time.Sleep(5 * time.Second)
	if err := startAllServicenodes(activator.Name); err != nil {
		return err
	}

	// TODO Check if wallets are reachable

	return nil
}

// getPortMap returns the port map configuration for the specified port.
func getPortMap(portExt, port, rpcExt, rpc, debugExt, debug string) nat.PortMap {
	ports := make(nat.PortMap)
	ports[nat.Port(port + "/tcp")] = []nat.PortBinding{
		{HostIP: "0.0.0.0", HostPort: portExt},
	}
	ports[nat.Port(rpc + "/tcp")] = []nat.PortBinding{
		{HostIP: "0.0.0.0", HostPort: rpcExt},
	}
	if debug != "" {
		ports[nat.Port(debug + "/tcp")] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: debugExt},
		}
	}
	return ports
}

// startServicenodes calls the servicenode start-all command on the activator.
func startAllServicenodes(name string) error {
	// Run servicenode start-all on activator
	cmdStartAll := blockRPCCommand(name, "servicenode start-all")
	if output, err := cmdStartAll.Output(); err != nil {
		return errors.Wrap(err, "Failed to run start-all on activator")
	} else {
		logrus.Debug(string(output))
	}
	return nil
}

// testBlocknetConfFile returns the path to the test blocknetdx.conf.
func testBlocknetConfFile() string {
	return path.Join(getConfigPath(), "localenv/blocknetdx.conf")
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
