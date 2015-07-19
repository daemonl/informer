package commands

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/daemonl/informer/shared"
)

type DiskCommand struct{}

func (dc *DiskCommand) GetCmd() *exec.Cmd {
	return exec.Command("df",
		"-x", "tmpfs",
		"-x", "devtmpfs",
		"--output=source,used,avail,target")
}

func (dc *DiskCommand) ParseOutput(stdout io.Reader) (interface{}, error) {
	disks := []shared.Disk{}
	s := ""
	fmt.Fscan(stdout, &s, &s, &s, &s, &s)
	for {
		d := shared.Disk{}
		_, err := fmt.Fscan(stdout,
			&d.Filesystem,
			&d.Used,
			&d.Available,
			&d.MountPoint)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err.Error())
			}
			break
		}
		disks = append(disks, d)
	}
	return disks, nil
}
