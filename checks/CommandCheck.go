package checks

import (
	"fmt"
	"os/exec"

	"github.com/daemonl/informer/reporter"
)

type CommandCheck struct {
	WorkingDirectory string   `xml:"dir"`
	Command          string   `xml:"cmd,attr"`
	Args             []string `xml:"arg"`
	Environment      Pairs    `xml:"env"`
}

func (t *CommandCheck) GetHash() string {
	return hashFromf("CMD:%s %s",
		t.WorkingDirectory,
		t.Command,
	)
}
func (t *CommandCheck) GetName() string {
	return t.Command
}

func (t *CommandCheck) RunCheck(reporter *reporter.Reporter) error {
	res := reporter.Report("Run command %s", t.Command)
	c := exec.Command(t.WorkingDirectory+"/"+t.Command, t.Args...)
	c.Dir = t.WorkingDirectory
	for _, envVar := range t.Environment {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
	}

	output, err := c.CombinedOutput()
	if err != nil {
		res.Fail("Error running %s: %s Dump: \n%s\n", t.Command, err.Error(), string(output))
	}
	return nil
}
