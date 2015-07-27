package checks

import (
	"github.com/daemonl/informer/reporter"
	"os/exec"
)

type CommandCheck struct {
	WorkingDirectory string `json:"workingDirectory"`
	Command          string `json:"command"`
}

func (t *CommandCheck) GetName() string {
	return t.Command
}

func (t *CommandCheck) RunCheck(reporter *reporter.Reporter) error {
	res := reporter.Report("Run command %s", t.Command)
	c := exec.Command(t.WorkingDirectory + "/" + t.Command)
	c.Dir = t.WorkingDirectory
	output, err := c.CombinedOutput()
	if err != nil {
		res.Fail("Error running %s: %s Dump: \n%s\n", t.Command, err.Error(), string(output))
	}
	return nil
}
