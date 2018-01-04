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

package util

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"runtime"

	"github.com/sirupsen/logrus"
)

// FileExists returns true if the path exists.
func FileExists(fPath string) bool {
	if _, err := os.Stat(fPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

// GetExecCmd returns the program name to be used with exec.Cmd instances.
func GetExecCmd() string {
	switch runtime.GOOS {
	case "windows":
		return "PowerShell"
	default:
		return "/bin/bash"
	}
}

// GetExecCmd returns the program name to be used with exec.Cmd instances.
func GetExecCmdSwitch() string {
	switch runtime.GOOS {
	case "windows":
		return "-Command"
	default:
		return "-c"
	}
}

// GetExecCmdConcat returns the program concat keyword for the current GOOS.
func GetExecCmdConcat() string {
	switch runtime.GOOS {
	case "windows":
		return ";"
	default:
		return "&&"
	}
}

// GetLocalIP returns the first local ip address found
var localIP string
func GetLocalIP() string {
	if localIP != "" {
		return localIP
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logrus.Error(err)
		return ""
	}
	defer conn.Close()
	localIP = conn.LocalAddr().(*net.UDPAddr).IP.String()
	return localIP
}

// Get24CIDR returns the /24 CIDR for the specified ip address.
func Get24CIDR(ip string) string {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(ip)
	if len(matches) <= 1 {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s.0/24", matches[1], matches[2], matches[3])
}

// CreateTar creates a tar file from a map of files.
func CreateTar(data map[string][]byte) (io.Reader, error) {
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	for path, datum := range data {
		hdr := tar.Header{
			Name: path,
			Mode: 0644,
			Size: int64(len(datum)),
		}
		if err := tw.WriteHeader(&hdr); err != nil {
			return nil, err
		}
		_, err := tw.Write(datum)
		if err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return &b, nil
}