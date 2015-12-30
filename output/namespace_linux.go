// Copyright (c) 2014-2016, b3log.org & hacpai.com
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

package output

import (
	"os/exec"
	"syscall"
)

func SetNamespace(cmd *exec.Cmd) {
	// XXX: keep move with Go 1.4 and later's

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	//cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET
	cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWUSER /*| syscall.CLONE_NEWNS*/ | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC /*| syscall.CLONE_NEWNET*/
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: 0,
		Gid: 0,
	}

	cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1001, Size: 1}}
	cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1001, Size: 1}}
}
