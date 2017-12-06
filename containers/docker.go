package containers

import (
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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// IsDockerInstalledAndRunning returns true if docker is installed. Returns false if docker
// is not installed and running or if error occurred when checking.
func IsDockerInstalledAndRunning() bool {
	var err error

	// Check if docker exists in path
	cmd := exec.Command("/bin/sh", "-c", `
		if [[ ! -z $(printf $(which docker)) ]]; then
			printf 'exists'
		else
			printf 'no'
		fi
	`)
	cmd.Stderr = os.Stderr
	var result []byte
	if result, err = cmd.Output(); err != nil {
		logrus.Error(err)
		return false
	}
	// Does docker exist?
	dockerExists := string(result) == "exists"

	// Check if docker is running
	cmdRu := exec.Command("/bin/sh", "-c", `
		if [[ $(docker ps) && $? == 0 ]]; then
			printf 'running'
		else
			printf 'no'
		fi
	`)
	cmdRu.Stderr = os.Stderr
	var resultR []byte
	if resultR, err = cmdRu.Output(); err != nil {
		logrus.Error(err)
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
		return err
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

// BuildImage builds image from path
func BuildImage(ctx context.Context, docker *client.Client, dir, dockerFile, imageName string) error {
	// Prep context in tar
	var includeFiles []string
	if err := filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Error(err)
			return nil
		}
		baseFilePath := strings.TrimLeft(strings.Replace(f, dir, "", 1), "/")
		baseFile := path.Base(f)
		if info.IsDir() || (baseFile != ".dockerignore" && strings.HasPrefix(baseFile, ".") || strings.Contains(f, ".git")) { // TODO Allow .dockerignore
			return nil
		}
		includeFiles = append(includeFiles, baseFilePath)
		return nil
	}); err != nil {
		return errors.Wrapf(err, "Failed to build image from source %s", dir)
	}
	tarOpts := &archive.TarOptions{
		Compression: archive.Uncompressed,
		IncludeFiles: includeFiles,
		ExcludePatterns: []string{".git*", "*.a", "*.o", ".*"},
	}
	dockerContext, err := archive.TarWithOptions(dir, tarOpts)
	if err != nil {
		return errors.Wrapf(err, "Failed to build image from source %s", dir)
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
	buildResponse, err := docker.ImageBuild(ctx, dockerContext, buildOpts)
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
	cmd := exec.Command("/bin/sh", "-c", `
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

	cm := exec.Command("/bin/sh", "-c", `
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
