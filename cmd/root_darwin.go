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
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// createConfigPath creates the configuration path and returns the path to the configuration file.
func createConfigPath() string {
	// Find runDir directory.
	runDir := getConfigPath()
	if err := os.MkdirAll(runDir, 0755); err != nil {
		logrus.Fatal(errors.Wrapf(err, "Failed to create the configuration directory at %s", runDir))
	}
	return runDir
}

// getConfigPath returns the configuration path.
func getConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Failed to get configuration path in $HOME/.dxregress %s", home))
		os.Exit(1)
	}
	return path.Join(home, ".dxregress")
}
