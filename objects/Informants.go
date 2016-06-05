package objects

import (
	"log"
	"os"
	"strings"

	"github.com/daemonl/informer/reporter"
)

type Informants struct {
	Methods []struct {
		Channel string `xml:"chan,attr"`
	} `xml:"inform"`
}

type InformEmail struct {
	Address string `xml:"address,attr"`
}

type InformChannel struct {
	Name   string        `xml:"name,attr"`
	Emails []InformEmail `xml:"email"`
	Apis   []InformAPI   `xml:"api"`
}

func (i *InformChannel) Call(core *Core, params InformParams) {
	for _, email := range i.Emails {
		core.Mailer.SendEmail(email.Address, params["subject"], params["body"])
	}
	for _, api := range i.Apis {
		DoApi(core, api.Name, params)
	}
}

func (core *Core) DoWarnings(r *reporter.Reporter, i *Informants) {

	warnings := r.CollectWarnings()
	if len(warnings) < 1 {
		return
	}

	body := strings.Join(warnings, "\n")
	hostname, _ := os.Hostname()
	body = body + hostname
	errId := r.ID

	params := map[string]string{
		"subject":  r.Name,
		"hostname": hostname,
		"body":     body,
		"id":       errId,
	}
	for _, method := range i.Methods {
		if len(method.Channel) > 0 {
			DoChannel(core, method.Channel, params)
		}
	}

}

func DoChannel(c *Core, name string, params InformParams) {
	var channel *InformChannel
	for _, thisChan := range c.Channels {
		if thisChan.Name == name {
			channel = &thisChan
			break
		}
	}
	if channel == nil {
		log.Printf("No chan named %s\n", name)
		return
	}
	channel.Call(c, params)
}

func DoApi(c *Core, name string, params InformParams) {
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
	api.Call(params)
}
