package objects

import (
	"log"
	"os"
	"strings"

	"github.com/daemonl/informer/reporter"
)

type Informants struct {
	Emails []struct {
		Address string `xml:"address,attr"`
	} `xml:"email"`
	Apis []struct {
		Name string `xml:"name,attr"`
	} `xml:"api"`
}

func (core *Core) DoWarnings(r *reporter.Reporter, i *Informants) {

	warnings := r.CollectWarnings()
	if len(warnings) < 1 {
		return
	}

	subject := r.Name
	body := strings.Join(warnings, "\n")
	hostname, _ := os.Hostname()
	body = body + hostname
	for _, email := range i.Emails {
		core.Mailer.SendEmail(email.Address, subject, body)
	}
	for _, api := range i.Apis {
		DoApi(core, api.Name, subject, body)
	}

}
func DoApi(c *Core, name string, subject string, body string) {
	var api *InformAPI
	for _, thisApi := range c.Apis {
		if thisApi.Name == name {
			api = &thisApi
			break
		}
	}
	if api == nil {
		log.Printf("No api call named %s\n", name)
		return
	}
	api.Call(subject, body)
}
