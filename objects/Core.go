package objects

import (
	"encoding/xml"
	"fmt"
	"sync"
	"time"

	"github.com/daemonl/informer/checks"
	"github.com/daemonl/informer/crosscheck"
	"github.com/daemonl/informer/reporter"
)

type Core struct {
	Mailer     *Mailer              `xml:"smtp"`
	Channels   []InformChannel      `xml:"chan"`
	Apis       []InformAPI          `xml:"api"`
	Admins     Informants           `xml:"admins"`
	Crosscheck *crosscheck.CXConfig `xml:"crosscheck"`
	Group
	Groups []Group `xml:"group"`
}

func (core Core) Run(runGroup string) {

	list := map[string][]Group{}

	for _, group := range core.Groups {
		// Matches "", which is 'unspecified'
		if runGroup == "all" || runGroup == group.RunGroup {
			_, ok := list[group.SyncGroup]
			if !ok {
				list[group.SyncGroup] = []Group{}
			}
			list[group.SyncGroup] = append(list[group.SyncGroup], group)
		}
	}

	wg := sync.WaitGroup{}
	for name, sg := range list {
		//fmt.Printf("Run sync %s - %d groups\n", name, len(sg))
		wg.Add(1)
		go func(name string, sg []Group) {
			defer wg.Done()
			start := time.Now().Unix()
			defer func() {
				duration := time.Now().Unix() - start
				fmt.Printf("%s took %d seconds\n", name, duration)
			}()
			for _, group := range sg {
				r := reporter.GetRoot(group.Name)
				r.ID = group.GetHash()
				for _, check := range group.Checks {
					err := check.RunCheck(r)
					if err != nil {
						r.AddError(err)
					}
				}
				r.DumpReport()
				core.DoWarnings(r, group.Informants)
			}

		}(name, sg)
	}
	wg.Wait()
}

type Group struct {
	Checks `xml:",any"`
	Informants
	RunGroup  string `xml:"run,attr"`
	Name      string `xml:"name,attr"`
	SyncGroup string `xml:"sync,attr"`
}

func (g *Group) GetHash() string {
	h := ""
	for _, check := range g.Checks {
		h += check.GetHash()
	}
	return h
}

type Checks []Check

type Checkable interface {
	RunCheck(*reporter.Reporter) error
	GetName() string
	GetHash() string
}

type Check struct {
	Checkable
}

func (c *Check) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {

	switch start.Name.Local {
	case "text":
		c.Checkable = &checks.TextCheck{}
	case "json":
		c.Checkable = &checks.JSONCheck{}
	case "log":
		c.Checkable = &checks.LogCheck{}
	case "command":
		c.Checkable = &checks.CommandCheck{}
	case "redirect":
		c.Checkable = &checks.RedirectCheck{}
	case "certificate":
		c.Checkable = &checks.CertificateCheck{}
	default:
		return fmt.Errorf("No checkable type %s", start.Name.Local)
	}

	return d.DecodeElement(c.Checkable, &start)
}
