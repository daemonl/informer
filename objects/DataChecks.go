package objects

import (
	"github.com/daemonl/informer/checks"
	"github.com/daemonl/informer/reporter"
)

type DataCheck struct {
	Informants
	Name       string             `xml:"name,attr"`
	LogChecks  []checks.LogCheck  `xml:"log"`
	JSONChecks []checks.JSONCheck `xml:"json"`
}

func (d DataCheck) RunChecks(r *reporter.Reporter) {
	for _, t := range d.LogChecks {
		err := t.RunCheck(r)
		if err != nil {
			r.AddError(err)
		}
	}
	for _, t := range d.JSONChecks {
		err := t.RunCheck(r)
		if err != nil {
			r.AddError(err)
		}
	}

}
