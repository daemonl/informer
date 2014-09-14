package checks

import (
	"fmt"
	"github.com/daemonl/informer/reporter"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type DiskCheck struct {
	HostName   string
	CheckDisks []DiskCheckDisk
}

type DiskCheckDisk struct {
	Filesystem string   `xml:"filesystem,attr"`
	MinBytes   *uint64  `xml:"minBytes"`
	MinPercent *float64 `xml:"minPercent"`
}

type DiskStatus struct {
	Filesystem string
	Used       uint64
	Available  uint64
	MountPoint string
}

var reSpace *regexp.Regexp = regexp.MustCompile(" +")

func (check *DiskCheck) RunCheck(reporter *reporter.Reporter) (err error) {

	disks, err := check.GetDisks()
	if err != nil {
		return err
	}

checkDisks:
	for _, checkDisk := range check.CheckDisks {
		for _, disk := range disks {
			if disk.Filesystem == checkDisk.Filesystem {

				if checkDisk.MinBytes != nil {
					res := reporter.Report("Check disk %s has > %d bytes free", disk.Filesystem, *checkDisk.MinBytes)
					if disk.Available < *checkDisk.MinBytes {
						res.Fail("%d bytes free", disk.Available)
					} else {
						res.Pass("%d bytes free", disk.Available)
					}

				}
				if checkDisk.MinPercent != nil {
					res := reporter.Report("Check disk %s has > %.2f%% free", disk.Filesystem, *checkDisk.MinPercent)
					fAvail := float64(disk.Available)
					fUsed := float64(disk.Used)
					fTotal := fUsed + fAvail
					availPercent := fAvail / fTotal * 100
					if availPercent < *checkDisk.MinPercent {

						res.Fail("%.2f%% free",
							availPercent)
					} else {
						res.Pass("%.2f%% free",
							availPercent)
					}

				}
				continue checkDisks
			}
		}
		res := reporter.Report("Check disk %s", checkDisk.Filesystem)
		res.Fail("Disk not found")
	}
	return

}

func (check *DiskCheck) GetDisks() (disks []*DiskStatus, err error) {

	// ssh chaos df
	args := []string{
		check.HostName,
		"df",
		"-P",
	}
	cmd := exec.Command("ssh", args...)
	resBytes, err := cmd.CombinedOutput()
	if err != nil {
		err := fmt.Errorf(string(resBytes))
		return disks, err
	}

	res := string(resBytes)

	lines := strings.Split(res, "\n")

	disks = make([]*DiskStatus, 0, len(lines)-1)

	for _, line := range lines[1:] {
		parts := reSpace.Split(line, -1)
		if len(parts) != 6 {
			if len(parts) > 1 {
				err = fmt.Errorf("Could not parse df line '%s'", line)
				return
			}
			continue
		}
		used, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return disks, fmt.Errorf("Could not parse df line '%s': %s", line, err.Error())
		}
		available, err := strconv.ParseUint(parts[3], 10, 64)
		if err != nil {
			return disks, fmt.Errorf("Could not parse df line '%s': %s", line, err.Error())
		}
		disk := &DiskStatus{
			Filesystem: parts[0],
			Used:       used,
			Available:  available,
			MountPoint: parts[5],
		}
		disks = append(disks, disk)
	}

	return
}
