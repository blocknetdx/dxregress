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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// GitApplyPatch applies the specified patch string to the specified codebase.
func GitApplyPatch(patch, patchPath, codebase string) error {
	// If the codebase doesn't exist return error
	if !FileExists(codebase) {
		return errors.New("Failed to find path " + codebase)
	}

	// Write patch to tmp
	if err := ioutil.WriteFile(patchPath, []byte(patch), 0755); err != nil {
		return errors.Wrapf(err, "Failed to write patch file to %s", patchPath)
	}

	// Apply patch to codebase (try revert patch if check fails)
	cmd := exec.Command("/bin/bash", "-c", strings.Replace(`
		check=$(git apply --check %s)
		if [[ $? != 0 ]]; then
			reset=$(git apply -R %s)
			if [[ $? != 0 ]]; then
				printf "Reverting patch failed, check codebase: git apply -R %s"
				exit 1
			fi
		fi
		apply=$(git apply %s)
		if [[ $? != 0 ]]; then
			printf "Patch failed for %s"
			exit 1
		else
			printf "Genesis patch applied"
		fi
	`, "%s", patchPath, -1))
	cmd.Dir = codebase
	if viper.GetBool("DEBUG") {
		cmd.Stderr = os.Stderr
	}

	result, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "Failed to apply genesis patch, possible conflict: %s", string(result))
	}
	logrus.Info(string(result))

	return nil
}

// GitRemovePatch removes the specified patch from the codebase.
func GitRemovePatch(patch, patchPath, codebase string) error {
	// If the codebase doesn't exist return error
	if !FileExists(codebase) {
		return errors.New("Failed to find path " + codebase)
	}

	// Write patch to tmp
	if err := ioutil.WriteFile(patchPath, []byte(patch), 0755); err != nil {
		return errors.Wrapf(err, "Failed to write patch file to %s", patchPath)
	}

	// Apply patch to codebase (try revert patch if check fails)
	cmd := exec.Command("/bin/bash", "-c", strings.Replace(`
		check=$(git apply --check %s)
		if [[ $? != 0 ]]; then
			reset=$(git apply -R %s)
			if [[ $? != 0 ]]; then
				printf "Reverting patch failed, check codebase: git apply -R %s"
				exit 1
			fi
		fi
	`, "%s", patchPath, -1))
	cmd.Dir = codebase

	result, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "Failed to remove genesis patch: %s", string(result))
	}
	if len(result) > 0 {
		logrus.Info(string(result))
	}

	return nil
}
