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

package chain

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BlocknetDX/dxregress/util"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	Activator = iota
	Sn1
	Sn2
)

type Node struct {
	ID           int
	ShortName    string
	Name         string
	Port         string
	RPCPort      string
	DebuggerPort string
	Ports        nat.PortMap
	CLI          string
	IsSnode      bool
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

// RPCCommand returns a command compatible with a running node.
func RPCCommand(name, exe, cmd string) *exec.Cmd {
	c := exec.Command(util.GetExecCmd(), util.GetExecCmdSwitch(), fmt.Sprintf("docker exec %s %s %s", name, exe, cmd))
	logrus.Debug(strings.Join(c.Args, " "))
	if viper.GetBool("DEBUG") {
		c.Stderr = os.Stderr
	}
	return c
}

// RPCCommands returns a command compatible with a node with multiple rpc calls.
func RPCCommands(name, exe string, cmds []string) *exec.Cmd {
	var cmdS string
	for i, c := range cmds {
		// Build the command
		cmdS += fmt.Sprintf("docker exec %s %s %s", name, exe, c)
		if i < len(cmds)-1 {
			cmdS += fmt.Sprintf(" %s ", util.GetExecCmdConcat())
		}
	}
	cmd := exec.Command(util.GetExecCmd(), util.GetExecCmdSwitch(), cmdS)
	if viper.GetBool("DEBUG") {
		cmd.Stderr = os.Stderr
	}
	return cmd
}

// RPCCommand returns a command compatible with a blocknet node.
func BlockRPCCommand(name, cmd string) *exec.Cmd {
	return RPCCommand(name, "blocknetdx-cli", cmd)
}

// BlockRPCCommands returns a command compatible with a blocknet node with multiple rpc calls.
func BlockRPCCommands(name string, cmds []string) *exec.Cmd {
	return RPCCommands(name, "blocknetdx-cli", cmds)
}

// StartServicenodesFrom calls the servicenode start-all command on the activator.
func StartServicenodesFrom(name string) error {
	// Run servicenode start-all on activator
	cmdStartAll := BlockRPCCommand(name, "servicenode start-all")
	if output, err := cmdStartAll.Output(); err != nil {
		return errors.Wrap(err, "Failed to run start-all on activator")
	} else {
		logrus.Debug(string(output))
	}
	return nil
}

// WaitForEnv will block for a maximum of N seconds until the local environment is ready. The
// getinfo rpc call is checked once every 2 seconds. This method returns if getinfo returns no
// error.
func WaitForEnv(parentContext context.Context, timeout time.Duration, nodes []Node) error {
	// Wait max 30 seconds for environment to provision
	ctx, cancel := context.WithTimeout(parentContext, timeout * time.Second)
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
					cmd := RPCCommand(node.Name, node.CLI, "getinfo")
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
			return errors.Wrap(ctx.Err(), "Timeout occurred while waiting for env to start up")
		}
	}

	return nil
}

// NodeForWallet returns a node configured with the specified xwallet.
func NodeForWallet(xwallet XWallet, containerPrefix string) Node {
	// TODO Add wallet debug port
	// Determine the debugger port (using port immediately after RPC port)
	var debugPort string
	//if port, err := strconv.Atoi(xwallet.RPCPort); err == nil {
	//	port++
	//	debugPort = strconv.Itoa(port)
	//}

	// Use the port as unique id
	nodeID, _ := strconv.Atoi(xwallet.Port)
	name := NodeContainerName(containerPrefix, xwallet.Name)
	if xwallet.BringOwn {
		name = "Virtual-" + xwallet.Name
	}
	return Node{
		nodeID,
		xwallet.Name,
		name,
		xwallet.Port,
		xwallet.RPCPort,
		debugPort,
		GetPortMap(xwallet.Port, xwallet.Port, xwallet.RPCPort, xwallet.RPCPort, debugPort, debugPort),
		xwallet.CLI,
		false,
	}
}

// GetPortMap returns the port map configuration for the specified port.
func GetPortMap(portExt, port, rpcExt, rpc, debugExt, debug string) nat.PortMap {
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

// NodeContainerName returns a valid dxregress container name with the specified prefix.
func NodeContainerName(prefix, name string) string {
	return prefix + name
}

// XWalletList returns a list of wallet names.
func XWalletList(wallets []XWallet) []string {
	var ws []string
	for _, wallet := range wallets {
		ws = append(ws, wallet.Name)
	}
	return ws
}

// NodeDockerfile returns the docker file.
func NodeDockerfile() string {
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
  && sleep 1 && ./autogen.sh \
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

// ServicenodeConf returns the servicenode conf file.
func ServicenodeConf(snodes []SNode) string {
	var r string
	for _, snode := range snodes {
		r += fmt.Sprintf("%s %s %s %s %s\n", snode.Alias, snode.IP, snode.Key, snode.CollateralID, snode.CollateralPos)
	}
	return r
}

// BlocknetdxConf returns a blocknetdx.conf with the specified parameters.
func BlocknetdxConf(currentNode int, nodes []Node, snodeKey string) string {
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
enableexchange=1
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

// XBridgeConf returns the xbridge configuration with the specified wallets.
func XBridgeConf(wallets []XWallet) string {
	conf := MAIN(XWalletList(wallets))
	for _, wallet := range wallets {
		conf += DefaultXConfig(wallet.Name, wallet.Version, wallet.Address, wallet.IP, wallet.RPCUser, wallet.RPCPass, wallet.BringOwn) + "\n\n"
	}
	return conf
}

// TestBlocknetConf
func TestBlocknetConf(containers []Node) string {
	return BlocknetdxConf(-1, containers, "")
}

// TestBlocknetConfFile returns the path to the test blocknetdx.conf.
func TestBlocknetConfFile(cpath string) string {
	return path.Join(cpath, "blocknetdx.conf")
}

// NodeForID returns the node data with the specified id.
func NodeForID(node int, nodes []Node) Node {
	for _, c := range nodes {
		if c.ID == node {
			return c
		}
	}
	return Node{}
}

// ServiceNodes returns the servicenodes.
func ServiceNodes(nodes []Node) []Node {
	var snodes []Node
	for _, c := range nodes {
		if c.IsSnode {
			snodes = append(snodes, c)
		}
	}
	return snodes
}