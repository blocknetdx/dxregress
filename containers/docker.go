package containers

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BlocknetDX/dxregress/chain"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// IsDockerInstalledAndRunning returns true if docker is installed. Returns false if docker
// is not installed and running or if error occurred when checking.
func IsDockerInstalledAndRunning() bool {
	var err error

	// Check if docker exists in path
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ ! -z $(printf $(which docker)) ]]; then
			printf 'exists'
		else
			printf 'no'
		fi
	`)
	cmd.Stderr = os.Stderr
	var result []byte
	if result, err = cmd.Output(); err != nil {
		logrus.Error(errors.Wrap(err, "Failed startup check: is docker installed?"))
		return false
	}
	// Does docker exist?
	dockerExists := string(result) == "exists"

	// Check if docker is running
	cmdRu := exec.Command("/bin/bash", "-c", `
		if [[ $(docker ps) && $? == 0 ]]; then
			printf 'running'
		else
			printf 'no'
		fi
	`)
	cmdRu.Stderr = os.Stderr
	var resultR []byte
	if resultR, err = cmdRu.Output(); err != nil {
		logrus.Error(errors.Wrap(err, "Failed startup check: is docker running?"))
		return false
	}
	// Is docker running
	dockerRunning := string(resultR) == "running"

	return dockerExists && dockerRunning
}

// CreateDockerfile at the specified path.
func CreateDockerfile(dockerFile, filePath string) error {
	if err := ioutil.WriteFile(filePath, []byte(dockerFile), 0755); err != nil {
		return errors.Wrapf(err, "Failed to write docker file to path %s", filePath)
	}
	return nil
}

// FindContainers returns containers with a name matching the specified regular expression.
func FindContainers(docker *client.Client, regex string) ([]types.Container, error) {
	// Find all containers matching name
	f := filters.NewArgs()
	f.Add("name", regex)
	containers, err := docker.ContainerList(context.TODO(), types.ContainerListOptions{Filters: f, All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// FindContainer returns the container matching the specified name.
func FindContainer(docker *client.Client, name string) types.Container {
	cs, err := FindContainers(docker, "^/" + name + "$")
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Failed to find container with name %s", name))
		return types.Container{}
	}
	if len(cs) < 1 {
		logrus.Error(errors.Wrapf(err, "Failed to find container with name %s", name))
		return types.Container{}
	}
	if len(cs) > 1 {
		logrus.Warn("Multiple containers matched [%s], returning first found", name)
	}
	return cs[0]
}

// StopAndRemove stops the container if it's already running and then removes the container.
func StopAndRemove(ctx context.Context, docker *client.Client, id string) error {
	result, err := docker.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}
	// If paused, resume before stopping
	if result.State.Paused {
		docker.ContainerStart(ctx, id, types.ContainerStartOptions{})
	}
	if result.State.Running {
		if e := StopContainer(ctx, docker, id); e != nil {
			return e
		}
	}
	if err = RemoveContainer(ctx, docker, id); err != nil {
		return err
	}
	return nil
}

// CreateAndStart creates and starts the container.
func CreateAndStart(ctx context.Context, docker *client.Client, image, name string, ports nat.PortMap) error {
	cfg := container.Config{
		Image: image,
		User: "root:root",
		Labels: map[string]string{
			"co.blocknet.dxregress": "true",
		},
	}
	hcfg := container.HostConfig{
		PortBindings: ports,
	}
	ncfg := network.NetworkingConfig{}
	result, err := docker.ContainerCreate(ctx, &cfg, &hcfg, &ncfg, name)
	if err != nil {
		return errors.Wrapf(err, "Failed to create %s container [%s]", name, image)
	}
	return docker.ContainerStart(ctx, result.ID, types.ContainerStartOptions{})
}

// StopContainer stops the container with the specified id.
func StopContainer(ctx context.Context, docker *client.Client, id string) error {
	dur := 30 * time.Second
	return docker.ContainerStop(ctx, id, &dur)
}

// RemoveContainer removes the container with the specified id.
func RemoveContainer(ctx context.Context, docker *client.Client, id string) error {
	return docker.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force:true})
}

// RestartContainers restarts all the containers matching the specified filter. The container
// will timeout after 30 seconds if the container's restart command hangs.
func RestartContainers(ctx context.Context, docker *client.Client, filter string) error {
	containers, err := FindContainers(docker, filter)
	if err != nil {
		return err
	}
	// Restart all nodes
	waitChan := make(chan error, 1)
	wg := new(sync.WaitGroup)
	for _, c := range containers {
		wg.Add(1)
		go func(c types.Container) {
			dur := time.Duration(30 * time.Second)
			if err := docker.ContainerRestart(ctx, c.ID, &dur); err != nil {
				logrus.Error(errors.Wrapf(err, "Failed to restart the container %s %s", c.Names[0], c.ID))
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		waitChan <- nil
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			return ctx.Err()
		}
	case <-waitChan:
	}

	return nil
}

// BuildImage builds image from path
func BuildImage(ctx context.Context, docker *client.Client, dir, dockerFile, imageName string) error {
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)
	// Prep context in tar
	if err := filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Error(err)
			return nil
		}
		baseFilePath := strings.TrimLeft(strings.Replace(f, dir, "", 1), "/")
		baseFile := path.Base(f)
		if info.IsDir() || (baseFile != ".dockerignore" && strings.HasPrefix(baseFile, ".") ||
			strings.Contains(f, ".git") || strings.HasSuffix(baseFile, ".o") ||
			strings.HasSuffix(baseFile, ".a")) {
			return nil
		}
		tarFileBytes, err := ioutil.ReadFile(f)
		hdr := &tar.Header{
			Name: baseFilePath,
			Mode: 0655,
			Size: int64(len(tarFileBytes)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			logrus.Error(err)
		}
		if _, err := tw.Write(tarFileBytes); err != nil {
			logrus.Error(err)
		}
		return nil
	}); err != nil {
		return errors.Wrapf(err, "Failed to build image from source %s", dir)
	}
	// Add default wallet file
	walletFile := chain.WalletData()
	if err := tw.WriteHeader(&tar.Header{
		Name: "wallet.dat",
		Mode: 0600,
		Size: int64(len(walletFile)),
	}); err != nil {
		logrus.Error(err)
	}
	if _, err := tw.Write(walletFile); err != nil {
		logrus.Error(err)
	}
	// Close tar file
	if err := tw.Close(); err != nil {
		return err
	}

	// Build and set labels
	labels := make(map[string]string)
	labels["co.blocknet.dxregress"] = "true"
	buildOpts := types.ImageBuildOptions{
		PullParent: true,
		Remove: true,
		Dockerfile: dockerFile,
		Labels: labels,
		Tags: []string{imageName},
	}
	buildResponse, err := docker.ImageBuild(ctx, tarBuf, buildOpts)
	if err != nil {
		return errors.Wrapf(err, "Failed to build image from source %s", dir)
	}
	defer buildResponse.Body.Close()

	// Read build response
	type JsonPacket struct {
		Stream string `json:"stream"`
		Status string `json:"status"`
	}
	js := json.NewDecoder(buildResponse.Body)
	for {
		var s JsonPacket
		if err := js.Decode(&s); err != nil {
			// Log error if non-EOF occurs
			if err != io.EOF {
				logrus.Error(err)
			}
			break
		}
		logrus.Infof("%s%s", strings.TrimSpace(s.Status), strings.TrimSpace(s.Stream))
	}

	return nil
}

// IsComposeInstalled returns true if docker compose is installed. Returns false if
// docker compose is not installed or if error occurred when checking.
func IsComposeInstalled() bool {
	cmd := exec.Command("/bin/bash", "-c", `
		if [[ -z $(printf $(which docker-compose)) ]]; then
			printf 'no'
		else
			printf 'yes'
		fi
	`)
	cmd.Stderr = os.Stderr

	var result []byte
	var err error
	if result, err = cmd.Output(); err != nil {
		return false
	}

	return string(result) == "yes"
}

// CreateTestNetwork creates the internal docker bridge network for use with servicenode regress testing.
// cidr must be specified in proper ipv4 format: 172.5.0.0/16
func CreateTestNetwork(cidr string) error {
	// Validate CIDR
	re := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)
	if !re.MatchString(cidr) {
		return errors.New(fmt.Sprintf("Bad CIDR %s: should be in format 0.0.0.0/0", cidr))
	}

	cm := exec.Command("/bin/bash", "-c", `
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
	cm.Stderr = os.Stderr

	// Run the command
	if err := cm.Run(); err != nil {
		return errors.Wrap(err, "Failed to create docker network")
	}

	return nil
}
