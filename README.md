# dxregress

Regression tester for BlocknetDX.

The tool uses a `.dxregress.yml` config at `$HOME/.dxregress.yml`.
Use `--config=` to load configurations of various setups.

# Installation

1) Install Prerequisites:
* Docker CE is required: https://www.docker.com/community-edition#/download
* Go v1.8+ (https://golang.org/dl/)

2) Install dxregress via go get:
```
go get github.com/BlocknetDX/dxregress
```

3) Add go/bin to $PATH (linux):
```
touch ~/.profile
echo "export GOPATH=/path/to/go" >> ~/.profile
echo "PATH=$GOPATH/bin:$PATH" >> ~/.profile
```

4) Run docker without sudo requires adding `$USER` to docker group (may need to logout):
```
sudo groupadd docker
sudo gpasswd -a $USER docker
```

# Commands 

## dxregress localenv up /path/to/src

The `localenv` command is used to test a development branch, including uncommitted code changes. This command applies a genesis patch to the codebase enabling regression testing in a local environment. Docker containers are used to facilitate communication between nodes. The local environment consists of 1 activator, 2 servicenodes and subsequent wallet nodes. The genesis patch is applied to the local codebase, as a result, building and running the code will result in the node joining the `localenv` environment. Normal debug operations are permitted as a result.

### Help
```
dxregress localenv -h
```

### Start localenv with SYS, MONA wallet support
```
dxregress localenv up -w=SYS,SRGU54nrCQWdKj4TUX1yT5PLabo9ESxJKt,test,testAbc -w=MONA,MRPfADFi2ohhmVqgnDrrUx7TuVCmiEY9bB,test,testAbc /path/to/codebase
```

### Add an existing wallet to localenv (requires explicitly specifying RPC IPv4 address)
```
dxregress localenv up -w=SYS,SRGU54nrCQWdKj4TUX1yT5PLabo9ESxJKt,test,testAbc,192.168.1.200 -w=MONA,MRPfADFi2ohhmVqgnDrrUx7TuVCmiEY9bB,test,testAbc,192.168.1.201 /path/to/codebase
```

### Stop localenv
```
dxregress localenv down /path/to/codebase
```

## Docker

Docker commands can be used to interact with the regression test environment.

### List all containers
```
docker ps
```

### Ask a wallet node for a new address
```
docker exec dxregress-localenv-MONA monacoin-cli getnewaddress
```

### Watch a wallet node's debug.log
```
docker exec dxregress-localenv-MONA tail -f /opt/blockchain/data/debug.log
```

### Watch a servicenode's debug.log
```
docker exec dxregress-localenv-sn1 tail -f /opt/blockchain/dxregress/testnet4/debug.log
```

# Copyright

```
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
```
