package objects

import (
	"github.com/daemonl/informer/checks"
	"github.com/daemonl/informer/reporter"
)

type Site struct {
	Name string `xml:"name,attr"`
	Informants
	RedirectTests []*checks.RedirectCheck `xml:"redirect"`
	SearchTests   []*checks.SearchCheck   `xml:"search"`
	CommandTests  []*checks.CommandCheck  `xml:"command"`
}

func (site Site) RunChecks(r *reporter.Reporter) {
	w := r
	for _, t := range site.RedirectTests {
		err := t.RunCheck(w)
		if err != nil {
			w.AddError(err)
		}
	}
	for _, t := range site.SearchTests {
		err := t.RunCheck(w)
		if err != nil {
			w.AddError(err)
		}
	}
	for _, t := range site.CommandTests {
		err := t.RunCheck(w)
		if err != nil {
			w.AddError(err)
		}
	}
}
