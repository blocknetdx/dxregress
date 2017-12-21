package chain

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Environment interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type EnvConfig struct {
	ConfigPath          string
	ContainerPrefix     string
	DefaultImage        string
	ContainerFilter     string
	ContainerFilterFunc func(filter string) string
	DockerFileName      string
	Activator           Node
	Nodes               []Node
	XWallets            []XWallet
}

// TestEnv is the default implementation for a test environment.
type TestEnv struct {
	config *EnvConfig
	docker *client.Client
	xwalletNodes []Node
}

// Start the environment.
func (env *TestEnv) Start(ctx context.Context) error {
	// Write test blocknetdx.conf file
	testLocalenvDir := path.Dir(TestBlocknetConfFile(env.config.ConfigPath))
	if err := os.MkdirAll(testLocalenvDir, 0775); err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", testLocalenvDir)
	}
	if err := ioutil.WriteFile(TestBlocknetConfFile(env.config.ConfigPath), []byte(TestBlocknetConf(env.config.Nodes)), 0644); err != nil {
		errors.Wrapf(err, "Failed to write blocknetdx.conf %s", TestBlocknetConfFile(env.config.ConfigPath))
	}

	// Stop all containers
	logrus.Info("Removing previous test containers...")
	if err := containers.StopAllContainers(ctx, env.docker, env.config.ContainerFilter, true); err != nil {
		logrus.Error(err)
	}

	// Start containers
	for _, c := range env.config.Nodes {
		if err := containers.CreateAndStart(ctx, env.docker, env.config.DefaultImage, c.Name, c.Ports); err != nil {
			return err
		}
		if c.DebuggerPort != "" {
			logrus.Infof("%s node running on %s, rpc on %s, gdb/lldb port on %s", c.Name, c.Port, c.RPCPort, c.DebuggerPort)
		} else {
			logrus.Infof("%s node running on %s, rpc on %s", c.Name, c.Port, c.RPCPort)
		}
	}

	// Start wallet containers
	for _, w := range env.config.XWallets {
		// Ignore BYOW nodes (bring your own wallet)
		if w.BringOwn {
			continue
		}
		// Create node from xwallet
		wc := NodeForWallet(w, env.config.ContainerPrefix)
		env.xwalletNodes = append(env.xwalletNodes, wc)
		if err := containers.CreateAndStart(ctx, env.docker, w.Container, wc.Name, wc.Ports); err != nil {
			return err
		}
		if wc.DebuggerPort != "" {
			logrus.Infof("%s node running on %s, rpc on %s, gdb/lldb port on %s", wc.Name, wc.Port, wc.RPCPort, wc.DebuggerPort)
		} else {
			logrus.Infof("%s node running on %s, rpc on %s", wc.Name, wc.Port, wc.RPCPort)
		}
	}

	logrus.Info("Waiting for nodes to be ready...")
	if err := WaitForEnv(ctx, 45, env.config.Nodes); err != nil {
		return err
	}

	// Setup blockchain
	if err := env.setupChain(ctx, env.docker); err != nil {
		return err
	}

	return nil
}

// Stop the environment, including performing necessary tear down.
func (env *TestEnv) Stop(ctx context.Context) error {
	if err := containers.StopAllContainers(ctx, env.docker, env.config.ContainerFilterFunc(""), false); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

// setupChain will setup the DX environment, copy all configuration files, test RPC connectivity.
func (env *TestEnv) setupChain(ctx context.Context, docker *client.Client) error {
	// activator wallet address: y5zBd8oLQSnTjChTUCfRieTAp5Z31bRwEV key: cQiWHyehhhsRFYadBpj5wQRU9HU23GtHSjyPY2hBLccHWeNq6iTY
	// sn1 alias address: y3DT9bZ69AjvdQFzYTCSpFgT9wJcRpHi7T key: cRdLcWroNyJPJ1BH4Q24pamDQtE3JNdm7tGQoD6mm9brqpYuX1dC
	// sn2 alias address: yF2E6wPBc1YosrGUMhgoet5zPat1A4Z87d key: cMn9aiQGBYqeRzRuTFAModv459UQNxGsXkgPSRQ1W7XwGdGCp1JB

	// Nodes
	activator := env.config.Activator
	snodes := ServiceNodes(env.config.Nodes)
	activatorC := containers.FindContainer(docker, activator.Name)

	// First import test address into alias and then generate test coin
	cmd := BlockRPCCommands(activator.Name, []string{"importprivkey cQiWHyehhhsRFYadBpj5wQRU9HU23GtHSjyPY2hBLccHWeNq6iTY coin", "setgenerate true 25"})
	if output, err := cmd.Output(); err != nil || string(output) == "" {
		return errors.Wrap(err, "Failed to generate first 25 blocks")
	} else {
		logrus.Debug(string(output))
	}

	// Import alias addresses
	cmd2 := BlockRPCCommands(activator.Name, []string{"importprivkey cRdLcWroNyJPJ1BH4Q24pamDQtE3JNdm7tGQoD6mm9brqpYuX1dC sn1", "importprivkey cMn9aiQGBYqeRzRuTFAModv459UQNxGsXkgPSRQ1W7XwGdGCp1JB sn2"})
	if output, err := cmd2.Output(); err != nil {
		return errors.Wrap(err, "Failed to import alias addresses")
	} else {
		logrus.Debug(string(output))
	}

	// Send 5k servicenode coin to each alias
	cmd3 := BlockRPCCommands(activator.Name, []string{"sendtoaddress y3DT9bZ69AjvdQFzYTCSpFgT9wJcRpHi7T 5000", "sendtoaddress yF2E6wPBc1YosrGUMhgoet5zPat1A4Z87d 5000"})
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
	cmd4 := BlockRPCCommands(activator.Name, cmd4S)
	if output, err := cmd4.Output(); err != nil {
		return errors.Wrap(err, "Failed to split coin")
	} else {
		logrus.Debug(string(output))
	}

	// Generate last 25 PoW blocks
	cmd5 := BlockRPCCommand(activator.Name, "setgenerate true 25")
	if output, err := cmd5.Output(); err != nil {
		return errors.Wrap(err, "Failed to generate blocks 25-50")
	} else {
		logrus.Debug(string(output))
	}

	// Obtain servicenode keys
	var keys []string
	for _, snode := range snodes {
		cmdSnode := BlockRPCCommand(snode.Name, "servicenode genkey")
		if output, err := cmdSnode.Output(); err != nil {
			return errors.Wrapf(err, "Failed to call genkey on %s", snode.Name)
		} else {
			keys = append(keys, strings.TrimSpace(string(output)))
		}
	}

	// Setup activator servicenode.conf
	type OutputsResponse struct {
		TxID  string `json:"txhash"`
		TxPos int    `json:"outputidx"`
	}
	cmd7 := BlockRPCCommand(activator.Name, "servicenode outputs")
	output, err := cmd7.Output()
	if err != nil {
		return errors.Wrap(err, "Failed to parse servicenode outputs")
	}
	var outputs []OutputsResponse
	if err := json.Unmarshal(output, &outputs); err != nil {
		return errors.Wrap(err, "Failed to parse servicenode outputs")
	}

	// Create servicenode specific data provider
	var servicenodes []SNode
	for i, snode := range snodes {
		ssn := SNode{
			ID: snode.ID,
			Alias: snode.ShortName,
			IP: snode.IP(),
			Key: keys[i],
			CollateralID: outputs[i].TxID,
			CollateralPos: strconv.Itoa(outputs[i].TxPos),
		}
		servicenodes = append(servicenodes, ssn)
	}

	// Max wait time for all commands below
	ctx, cancel := context.WithTimeout(context.Background(), 180 * time.Second)
	defer cancel()

	// Generate activator servicenode.conf
	snConf := ServicenodeConf(servicenodes)
	// Copy activator servicenode.conf
	if servicenodeConf, err := util.CreateTar(map[string][]byte{"servicenode.conf": []byte(snConf)}); err == nil {
		if err := docker.CopyToContainer(ctx, activatorC.ID, "/opt/blockchain/dxregress/testnet4/", servicenodeConf, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write servicenode.conf to activator")
		}
	} else {
		return errors.Wrap(err, "Failed to write servicenode.conf to activator")
	}

	// Update activator blocknetdx.conf
	blocknetConfActivator := BlocknetdxConf(Activator, env.config.Nodes, "")
	if bufActivator, err := util.CreateTar(map[string][]byte{"blocknetdx.conf": []byte(blocknetConfActivator)}); err == nil {
		if err := docker.CopyToContainer(ctx, activatorC.ID, "/opt/blockchain/config/", bufActivator, types.CopyToContainerOptions{}); err != nil {
			return errors.Wrap(err, "Failed to write blocknetdx.conf to activator")
		}
	} else {
		return errors.Wrap(err, "Failed to write blocknetdx.conf to activator")
	}

	// Copy config files to servicenodes
	xbridgeConfSnode := XBridgeConf(env.config.XWallets)
	for _, ssn := range servicenodes {
		sn := NodeForID(ssn.ID, snodes)
		ssnC := containers.FindContainer(env.docker, sn.Name)

		// Update servicenodes blocknetdx.conf
		blocknetConfSn := BlocknetdxConf(ssn.ID, env.config.Nodes, ssn.Key)
		if bufSn, err := util.CreateTar(map[string][]byte{"blocknetdx.conf": []byte(blocknetConfSn)}); err == nil {
			if err := docker.CopyToContainer(ctx, ssnC.ID, "/opt/blockchain/config/", bufSn, types.CopyToContainerOptions{}); err != nil {
				return errors.Wrapf(err, "Failed to write blocknetdx.conf to %s", sn.ShortName)
			}
		} else {
			return errors.Wrapf(err, "Failed to write blocknetdx.conf to %s", sn.ShortName)
		}

		// Write servicenodes xbridge.conf
		if bufSn, err := util.CreateTar(map[string][]byte{"xbridge.conf": []byte(xbridgeConfSnode)}); err == nil {
			if err := docker.CopyToContainer(ctx, ssnC.ID, "/opt/blockchain/dxregress/", bufSn, types.CopyToContainerOptions{}); err != nil {
				return errors.Wrapf(err, "Failed to write xbridge.conf to %s", sn.ShortName)
			}
		} else {
			return errors.Wrapf(err, "Failed to copy xbridge.conf to %s", sn.ShortName)
		}
	}

	// Stop activator
	if err := containers.StopContainer(ctx, docker, activatorC.ID); err != nil {
		return err
	}

	// Restart all nodes except for wallet nodes
	if err := containers.RestartContainers(ctx, docker, env.config.ContainerFilterFunc("sn")); err != nil {
		return err
	}
	if err := containers.RestartContainers(ctx, docker, env.config.ContainerFilterFunc("act")); err != nil {
		return err
	}

	// Wait for nodes to be ready
	logrus.Info("Waiting for nodes and wallets to be ready...")
	allContainers := append(env.config.Nodes, env.xwalletNodes...)
	if err := WaitForEnv(ctx, 45, allContainers); err != nil {
		return err
	}

	// Start servicenodes
	if err := StartServicenodesFrom(activator.Name); err != nil {
		return err
	}

	// Wait before restarting staker
	time.Sleep(10 * time.Second)

	// Restart the activator to trigger staking
	if err := containers.RestartContainers(ctx, docker, env.config.ContainerFilterFunc("act")); err != nil {
		return err
	}

	// Wait for activator to be ready
	logrus.Info("Waiting for activator to be ready...")
	if err := WaitForEnv(ctx, 45, []Node{activator}); err != nil {
		return err
	}

	// Call start servicenodes a second time to make sure they're started
	// Wait before re-running snode command
	time.Sleep(5 * time.Second)
	if err := StartServicenodesFrom(activator.Name); err != nil {
		return err
	}

	// TODO Check if wallets are reachable

	return nil
}

// NewTestEnvironment creates a new test environment instance.
func NewTestEnv(config *EnvConfig, docker *client.Client) *TestEnv {
	env := new(TestEnv)
	env.config = config
	env.docker = docker
	return env
}
