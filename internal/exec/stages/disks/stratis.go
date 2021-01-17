// Copyright Red Hat, Inc.
//
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

// The storage stage is responsible for partitioning disks, creating RAID
// arrays, formatting partitions, writing files, writing systemd units, and
// writing network units.
// createRaids creates the raid arrays described in config.Storage.Raid.

package disks

import (
	"fmt"
	"os/exec"

	"github.com/coreos/ignition/v2/config/v3_3_experimental/types"
	"github.com/coreos/ignition/v2/internal/distro"
	"github.com/coreos/ignition/v2/internal/exec/util"
)

func (s stage) createStratisFilesystems(config types.Config) error {
	if len(config.Storage.StratisPools) == 0 && len(config.Storage.StratisFilesystems) == 0 {
		return nil
	}
	s.Logger.PushPrefix("createStratis")
	defer s.Logger.PopPrefix()

	devs := []string{}
	for _, pool := range config.Storage.StratisPools {
		for _, dev := range pool.Devices {
			devs = append(devs, string(dev))
		}
	}

	if err := s.waitOnDevicesAndCreateAliases(devs, "stratis"); err != nil {
		return err
	}

	for _, pool := range config.Storage.StratisPools {
		args := []string{"pool", "create"}

		for _, dev := range pool.Devices {
			args = append(args, util.DeviceAlias(string(dev)))
		}

		if _, err := s.Logger.LogCmd(
			exec.Command(distro.MdadmCmd(), args...),
			"creating Stratis pool %q", pool.Name,
		); err != nil {
			return fmt.Errorf("stratis failed: %v", err)
		}
	}

	for _, fs := range config.Storage.StratisFilesystems {
		args := []string{"filesystem", "create", fs.PoolName, fs.Name}

		if _, err := s.Logger.LogCmd(
			exec.Command(distro.StratisCmd(), args...),
			"creating Stratis filesystem %q", fs.Name,
		); err != nil {
			return fmt.Errorf("stratis failed: %v", err)
		}
	}

	return nil
}
