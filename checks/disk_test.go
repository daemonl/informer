package checks

import (
	"fmt"
	"testing"

	"github.com/daemonl/informer/warner"
)

func TestDisk(t *testing.T) {

	w := warner.GetWarner()

	b1 := uint64(6000000)
	b2 := uint64(1000)

	p1 := float64(90)
	p2 := float64(10)

	dc := &DiskCheck{
		Hostname: "chaos",
		CheckDisks: []DiskCheckDisk{
			DiskCheckDisk{
				Filesystem: "IDONTEXIST",
			},
			DiskCheckDisk{
				Filesystem: "/dev/xvda1",
				MinBytes:   &b1,
			},
			DiskCheckDisk{
				Filesystem: "/dev/xvda1",
				MinBytes:   &b2,
			},
			DiskCheckDisk{
				Filesystem: "/dev/xvda1",
				MinPercent: &p1,
			},
			DiskCheckDisk{
				Filesystem: "/dev/xvda1",
				MinPercent: &p2,
			},
		},
	}

	err := dc.RunCheck(w)
	if err != nil {
		fmt.Println(err.Error())
	}

	w.Dump()

}
