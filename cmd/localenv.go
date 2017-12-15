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
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/BlocknetDX/dxregress/containers"
	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const localenvPrefix = "dxregress-localenv-"
const genesisPatchFile = "dxregress.patch"
const dockerFilePath = "Dockerfile-dxregress"
const containerImage = "blocknetdx/dxregress:localenv"

var codedir string

const (
	Activator = iota
	Sn1
	Sn2
)

// localenv containers
type Node struct {
	ID           int
	ShortName    string
	Name         string
	Port         string
	RPCPort      string
	DebuggerPort string
	Ports        nat.PortMap
}

func (node Node) IP() string {
	return util.GetLocalIP() + ":" + node.Port
}

type SNode struct {
	ID            int
	Alias         string
	IP            string
	Key           string
	CollateralID  string
	CollateralPos string
}

var localContainers = []Node{
	{Activator, "activator", dxregressContainerName("activator"), "41477", "41427", "41487", getPortMap("41477", "41427", "41487")},
	{Sn1, "sn1", dxregressContainerName("sn1"), "41478", "41428", "41488", getPortMap("41478", "41428", "41488")},
	{Sn2, "sn2", dxregressContainerName("sn2"), "41479", "41429", "41489", getPortMap("41479", "41429", "41489")},
}

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

// stopAllLocalEnvContainers stops the existing localenv containers.
func stopAllLocalEnvContainers(ctx context.Context, docker *client.Client, suppressLogs bool) error {
	containerList, err := containers.FindContainers(docker, localEnvContainerFilter(""))
	if err != nil {
		return err
	}
	if len(containerList) == 0 {
		logrus.Info("No localenv containers")
		return nil
	}

	// Stop containers in parallel
	wg := new(sync.WaitGroup)
	for _, c := range containerList {
		wg.Add(1)
		go func(c types.Container) {
			name := c.Names[0]
			if !suppressLogs {
				logrus.Infof("Removing localenv container %s, please wait...", name)
			}
			if err := containers.StopAndRemove(ctx, docker, c.ID); err != nil {
				logrus.Errorf("Failed to remove %s: %s", name, err.Error())
			} else if !suppressLogs {
				logrus.Infof("Removed %s", name)
			}
			wg.Done()
		}(c)
	}

	waitChan := make(chan bool, 1)
	go func() {
		wg.Wait()
		waitChan <- true
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waitChan:
	}

	return nil
}

// rpcParams returns the connection parameters.
func rpcParams(node int) string {
	return "-rpcuser=localenv -rpcpassword=test -rpcport=" + rpcPortForNode(node)
}

// rpcCommand returns a command compatible with the activator node.
func rpcCommand(node int, cmd string) *exec.Cmd {
	c := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s %s %s", path.Join(codedir, "src/blocknetdx-cli"), rpcParams(node), cmd))
	logrus.Debug(strings.Join(c.Args, " "))
	if viper.GetBool("DEBUG") {
		c.Stderr = os.Stderr
	}
	return c
}

// rpcCommands returns a command compatible with the activator node that includes multiple rpc commands.
func rpcCommands(node int, cmds []string) *exec.Cmd {
	var cmdS string
	for i, c := range cmds {
		// Build the command
		cmdS += fmt.Sprintf("%s %s %s ", path.Join(codedir, "src/blocknetdx-cli"), rpcParams(node), c)
		if i < len(cmds)-1 {
			cmdS += "&& "
		}
	}
	cmd := exec.Command("/bin/sh", "-c", cmdS)
	if viper.GetBool("DEBUG") {
		cmd.Stderr = os.Stderr
	}
	return cmd
}

// rpcPortForNode returns the rpc port for the specified node.
func rpcPortForNode(nodeID int) string {
	for _, node := range localContainers {
		if node.ID == nodeID {
			return node.RPCPort
		}
	}
	return ""
}

// localContainerForNode returns the node data with the specified id.
func localContainerForNode(node int) Node {
	for _, c := range localContainers {
		if c.ID == node {
			return c
		}
	}
	return Node{}
}

// dockerFile returns the docker file.
func dockerFile() string {
	return `FROM ubuntu:trusty

ARG cores=` + fmt.Sprintf("%d", runtime.NumCPU()) + `
ENV ecores=$cores

RUN apt update \
  && apt install -y --no-install-recommends \
     software-properties-common \
     ca-certificates \
     wget curl git python vim \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN add-apt-repository ppa:bitcoin/bitcoin \
  && apt update \
  && apt install -y --no-install-recommends \
     build-essential libtool autotools-dev bsdmainutils \
     libevent-dev autoconf automake pkg-config libssl-dev \
     libboost-system-dev libboost-filesystem-dev libboost-chrono-dev \
     libboost-program-options-dev libboost-test-dev libboost-thread-dev \
     libdb4.8-dev libdb4.8++-dev libgmp-dev libminiupnpc-dev libzmq3-dev \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Build berkeleydb4.8
RUN mkdir -p /tmp/berkeley \
  && cd /tmp/berkeley \
  && wget 'http://download.oracle.com/berkeley-db/db-4.8.30.NC.tar.gz' \
  && [ "$(printf '12edc0df75bf9abd7f82f821795bcee50f42cb2e5f76a6a281b85732798364ef db-4.8.30.NC.tar.gz' | sha256sum -c)" = "db-4.8.30.NC.tar.gz: OK" ] || $(echo "Berkeley DB 4.8 sha256sum failed"; exit 1) \
  && tar -xzvf db-4.8.30.NC.tar.gz \
  && cd db-4.8.30.NC/build_unix/ \
  && ../dist/configure --enable-cxx --disable-shared --with-pic --prefix=/tmp/berkeley \
  && make install

COPY . /opt/blocknetdx/BlockDX/

# Build source
RUN mkdir -p /opt/blockchain/config \
  && mkdir -p /opt/blockchain/dxregress/testnet4 \
  && ln -s /opt/blockchain/config /root/.blocknetdx \
  && cp /opt/blocknetdx/BlockDX/wallet.dat /opt/blockchain/dxregress/testnet4/wallet.dat

RUN cd /opt/blocknetdx/BlockDX \
  && chmod +x ./autogen.sh \
  && ./autogen.sh \
  && ./configure LDFLAGS="-L/tmp/berkeley/lib/" CPPFLAGS="-I/tmp/berkeley/include/" --without-gui --enable-debug --enable-tests=0 \
  && make clean \
  && make -j$ecores \
  && make install \
  && rm -rf /opt/blocknetdx/ /tmp/berkeley/*

# Write default blocknetdx.conf
RUN echo "datadir=/opt/blockchain/dxregress \n\
                                            \n\
testnet=1                                   \n\
dbcache=256                                 \n\
maxmempool=512                              \n\
                                            \n\
port=41476                                  \n\
rpcport=41419                               \n\
                                            \n\
listen=1                                    \n\
server=1                                    \n\
maxconnections=10                           \n\
logtimestamps=1                             \n\
logips=1                                    \n\
                                            \n\
rpcuser=localenv                            \n\
rpcpassword=test                            \n\
rpcallowip=0.0.0.0/0                        \n\
rpctimeout=15                               \n\
rpcclienttimeout=15" > /opt/blockchain/config/blocknetdx.conf

WORKDIR /opt/blockchain/
VOLUME ["/opt/blockchain/config", "/opt/blockchain/dxregress"]

# Testnet Port, RPC, GDB Remote Debug
EXPOSE 41476 41419 41475

CMD ["blocknetdxd", "-daemon=0", "-testnet=1", "-conf=/root/.blocknetdx/blocknetdx.conf"]
`
}

// servicenodeConf returns the servicenode conf file.
func servicenodeConf(snodes []SNode) string {
	var r string
	for _, snode := range snodes {
		r += fmt.Sprintf("%s %s %s %s %s\n", snode.Alias, snode.IP, snode.Key, snode.CollateralID, snode.CollateralPos)
	}
	return r
}

// blocknetdxConf returns a blocknetdx.conf with the specified parameters.
func blocknetdxConf(currentNode int, nodes []Node, snodeKey string) string {
	base := `datadir=/opt/blockchain/dxregress
testnet=1
dbcache=256
maxmempool=512

port=41476
rpcport=41419

listen=1
server=1
logtimestamps=1
logips=1

rpcuser=localenv
rpcpassword=test
rpcallowip=0.0.0.0/0
rpctimeout=15
rpcclienttimeout=15

`
	localIP := util.GetLocalIP()
	base += `whitelist=0.0.0.0/0
`

	var cnode Node
	for _, node := range nodes {
		// do not addnode to self
		if node.ID == currentNode {
			cnode = node
			continue
		}
		base += fmt.Sprintf("connect=%s:%s\n", localIP, node.Port)
	}

	// Add servicenode config
	if snodeKey != "" {
		base += `
staking=0
servicenode=1
servicenodeaddr=` + fmt.Sprintf("%s:%s", localIP, cnode.Port) + `
servicenodeprivkey=` + snodeKey + `
`
	} else { // support staking on non-servicenode clients
		base += `staking=1
`
	}
	return base
}

// xbridgeConf returns the xbridge configuration with the specified wallets.
func xbridgeConf(wallets []string) string {
	ip := util.GetLocalIP()
	conf := chain.MAIN(wallets)
	for _, wallet := range wallets {
		conf += chain.DefaultXConfig(wallet, "", ip, "localenv", "test") + "\n"
	}
	return conf
}

// dxregressContainerName returns a valid dxregress container name.
func dxregressContainerName(name string) string {
	return localenvPrefix + name
}

// testBlocknetConf
func testBlocknetConf() string {
	return blocknetdxConf(-1, localContainers, "")
}
