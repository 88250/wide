package output

import (
	"os"
	"sync"

	"github.com/b3log/wide/session"
	"github.com/golang/glog"
)

// Type of process set.
type procs map[string][]*os.Process

// Processse of all users.
//
// <sid, []*os.Process>
var processes = procs{}

// Exclusive lock.
var mutex sync.Mutex

// add adds the specified process to the user process set.
func (procs *procs) add(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.Id
	userProcesses := (*procs)[sid]

	userProcesses = append(userProcesses, proc)
	(*procs)[sid] = userProcesses

	// bind process with wide session
	wSession.SetProcesses(userProcesses)

	glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))
}

// remove removes the specified process from the user process set.
func (procs *procs) remove(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.Id

	userProcesses := (*procs)[sid]

	var newProcesses []*os.Process
	for i, p := range userProcesses {
		if p.Pid == proc.Pid {
			newProcesses = append(userProcesses[:i], userProcesses[i+1:]...) // remove it
			(*procs)[sid] = newProcesses

			// bind process with wide session
			wSession.SetProcesses(newProcesses)

			glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))

			return
		}
	}
}

// kill kills a process specified by the given pid.
func (procs *procs) kill(wSession *session.WideSession, pid int) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.Id

	userProcesses := (*procs)[sid]

	for i, p := range userProcesses {
		if p.Pid == pid {
			if err := p.Kill(); nil != err {
				glog.Error("Kill a process [pid=%d] of session [%s] failed [error=%v]", pid, sid, err)
			} else {
				var newProcesses []*os.Process

				newProcesses = append(userProcesses[:i], userProcesses[i+1:]...)
				(*procs)[sid] = newProcesses

				// bind process with wide session
				wSession.SetProcesses(newProcesses)

				glog.V(3).Infof("Killed a process [pid=%d] of session [%s]", pid, sid)
			}

			return
		}
	}
}
