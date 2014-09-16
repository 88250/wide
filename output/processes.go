package output

import (
	"os"

	"github.com/golang/glog"
)

// 所有用户正在运行的程序进程集.
// <sid, []*os.Process>
type procs map[string][]*os.Process

var processes = procs{}

// 添加用户执行进程.
func (procs *procs) add(sid string, proc *os.Process) {
	userProcesses := (*procs)[sid]

	userProcesses = append(userProcesses, proc)
	(*procs)[sid] = userProcesses

	glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))
}

// 移除用户执行进程.
func (procs *procs) remove(sid string, proc *os.Process) {
	userProcesses := (*procs)[sid]

	var newProcesses []*os.Process
	for i, p := range userProcesses {
		if p.Pid == proc.Pid {
			newProcesses = append(userProcesses[:i], userProcesses[i+1:]...)
			(*procs)[sid] = newProcesses

			glog.V(3).Infof("Session [%s] has [%d] processes", sid, len((*procs)[sid]))

			return
		}
	}
}

// 结束用户正在执行的进程.
func (procs *procs) kill(sid string, pid int) {
	pros := (*procs)[sid]

	for _, p := range pros {
		if p.Pid == pid {
			if err := p.Kill(); nil != err {
				glog.Error("Kill a process [pid=%d] of session [%s] failed [error=%v]", pid, sid, err)
			} else {
				glog.V(3).Infof("Killed a process [pid=%d] of session [%s]", pid, sid)
			}
		}
	}
}
