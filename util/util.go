package util

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
)

// FileExists returns true if the path exists.
func FileExists(fPath string) bool {
	if _, err := os.Stat(fPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

// GetLocalIP returns the first local ip address found
func GetLocalIP() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addresses {
		// ignore loopback
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP.String()
			}
		}
	}
	return ""
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