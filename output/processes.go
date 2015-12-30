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
	"os"
	"sync"

	"github.com/b3log/wide/session"
)

// Type of process set.
type procs map[string][]*os.Process

// Processse of all users.
//
// <sid, []*os.Process>
var Processes = procs{}

// Exclusive lock.
var mutex sync.Mutex

// Add adds the specified process to the user process set.
func (procs *procs) Add(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.ID
	userProcesses := (*procs)[sid]

	userProcesses = append(userProcesses, proc)
	(*procs)[sid] = userProcesses

	// bind process with wide session
	wSession.SetProcesses(userProcesses)

	logger.Tracef("Session [%s] has [%d] processes", sid, len((*procs)[sid]))
}

// Remove removes the specified process from the user process set.
func (procs *procs) Remove(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.ID

	userProcesses := (*procs)[sid]

	var newProcesses []*os.Process
	for i, p := range userProcesses {
		if p.Pid == proc.Pid {
			newProcesses = append(userProcesses[:i], userProcesses[i+1:]...) // remove it
			(*procs)[sid] = newProcesses

			// bind process with wide session
			wSession.SetProcesses(newProcesses)

			logger.Tracef("Session [%s] has [%d] processes", sid, len((*procs)[sid]))

			return
		}
	}
}

// Kill kills a process specified by the given pid.
func (procs *procs) Kill(wSession *session.WideSession, pid int) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.ID

	userProcesses := (*procs)[sid]

	for i, p := range userProcesses {
		if p.Pid == pid {
			if err := p.Kill(); nil != err {
				logger.Errorf("Kill a process [pid=%d] of user [%s, %s] failed [error=%v]", pid, wSession.Username, sid, err)
			} else {
				var newProcesses []*os.Process

				newProcesses = append(userProcesses[:i], userProcesses[i+1:]...)
				(*procs)[sid] = newProcesses

				// bind process with wide session
				wSession.SetProcesses(newProcesses)

				logger.Debugf("Killed a process [pid=%d] of user [%s, %s]", pid, wSession.Username, sid)
			}

			return
		}
	}
}
