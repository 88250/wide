package output

import (
	"os"
	"sync"

	"github.com/b3log/wide/session"
	"github.com/golang/glog"
)

// 所有用户正在运行的程序进程集.
// <sid, []*os.Process>
type procs map[string][]*os.Process

var processes = procs{}

// 排它锁，防止并发修改.
var mutex sync.Mutex

// 添加用户执行进程.
func (procs *procs) add(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.Id
	userProcesses := (*procs)[sid]

	userProcesses = append(userProcesses, proc)
	(*procs)[sid] = userProcesses

	// 会话关联进程
	wSession.SetProcesses(userProcesses)

	glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))
}

// 移除用户执行进程.
func (procs *procs) remove(wSession *session.WideSession, proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()

	sid := wSession.Id

	userProcesses := (*procs)[sid]

	var newProcesses []*os.Process
	for i, p := range userProcesses {
		if p.Pid == proc.Pid {
			newProcesses = append(userProcesses[:i], userProcesses[i+1:]...)
			(*procs)[sid] = newProcesses

			// 会话关联进程
			wSession.SetProcesses(newProcesses)

			glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))

			return
		}
	}
}

// 结束用户正在执行的进程.
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

				// 会话关联进程
				wSession.SetProcesses(newProcesses)

				glog.V(3).Infof("Killed a process [pid=%d] of session [%s]", pid, sid)
			}
		}
	}
}
