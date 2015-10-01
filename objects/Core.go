package objects

import (
	"encoding/xml"
	"fmt"

	"github.com/daemonl/informer/checks"
	"github.com/daemonl/informer/reporter"
)

type Core struct {
	Mailer *Mailer     `xml:"smtp"`
	Apis   []InformAPI `xml:"api"`
	Group
	Groups []Group `xml:"group"`
}

type Group struct {
	Checks `xml:",any"`
	Informants
	RunGroup  string `xml:"run,attr"`
	Name      string `xml:"name,attr"`
	SyncGroup string `xml:"sync,attr"`
}

type Checks []Check

type Checkable interface {
	RunCheck(*reporter.Reporter) error
	GetName() string
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
