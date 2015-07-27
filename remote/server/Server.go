package server

import (
	"os/exec"

	"github.com/daemonl/informer/reporter"
)

type Server struct {
	Name string `xml:"name,attr"`
	SSH
	Disks []Disk `xml:"disk"`
}

type SSH struct {
	HostName string `xml:"hostName"`
	KeyFile  string `xml:"keyFile"`
	Username string `xml:"username"`
}

func (s *Server) RunChecks(r *reporter.Reporter) {
	if s.SSH.HostName == "" {
		s.SSH.HostName = s.Name
	}
	err := s.RunDiskCheck(r)
	if err != nil {
		r.AddError(err)
	}
}

func (ssh *SSH) RPC(cmd string, remoteArgs ...string) *exec.Cmd {

	args := []string{
		"-o",
		"LogLevel Error",
	}

	//-i keyfile
	if len(ssh.KeyFile) > 0 {
		args = append(args, "-i", ssh.KeyFile)
	}
	//user@host
	hostPart := ssh.HostName
	if len(ssh.Username) > 0 {
		hostPart = ssh.Username + "@" + ssh.HostName
	}

	args = append(args, hostPart)
	args = append(args, cmd)
	args = append(args, remoteArgs...)
	return exec.Command("ssh", args...)
}
