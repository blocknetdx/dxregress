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

package containers

import (
	"os"
	"os/exec"
)

// cmdDockerExists determines if docker exists. Prints "exists" or "no".
func cmdDockerExists() *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ ! -z $(printf $(which docker)) ]]; then
			printf 'exists'
		else
			printf 'no'
		fi
	`)
	cmd.Stderr = os.Stderr
	return cmd
}

// cmdDockerIsRunning determines if docker is running. Prints "running" or "no".
func cmdDockerIsRunning() *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ $(docker ps) && $? == 0 ]]; then
			printf 'running'
		else
			printf 'no'
		fi
	`)
	cmd.Stderr = os.Stderr
	return cmd
}

// cmdComposeIsInstalled determines if docker-compose exists. Prints "yes" or "no".
func cmdComposeIsInstalled() *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ -z $(printf $(which docker-compose)) ]]; then
			printf 'no'
		else
			printf 'yes'
		fi
	`)
	cmd.Stderr = os.Stderr
	return cmd
}

// cmdCreateDockerNetwork creates the blocknet docker network.
func cmdCreateDockerNetwork() *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ -z $(docker network ls -qf name=blocknet) ]]; then
			docker network create --subnet 172.5.0.0/16 --gateway 172.5.0.1 \
				-o "com.docker.network.bridge.enable_icc"="true" \
				-o "com.docker.network.bridge.enable_ip_masquerade"="true" \
				-o "com.docker.network.bridge.host_binding_ipv4"="127.0.0.1" \
				-o "com.docker.network.driver.mtu"="1500" \
				-o "com.docker.network.bridge.name"="blocknet0" blocknet
			echo "Regression test network created"
		fi
	`)
	cmd.Stderr = os.Stderr
	return cmd
}