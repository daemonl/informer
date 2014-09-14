package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"github.com/daemonl/informer/checks"
	"github.com/daemonl/informer/reporter"

	"sync"
)

var configFilename string

func init() {
	flag.StringVar(&configFilename, "config", "/etc/informer/sites.xml", "Config file")
}

type Err struct {
	Message string
}

func (e *Err) Error() string {
	return e.Message
}

func ErrF(format string, parameters ...interface{}) error {
	return &Err{Message: fmt.Sprintf(format, parameters...)}
}

type SmtpConfig struct {
	ServerAddress string `xml:"serverAddress"`
	ServerPort    string `xml:"serverPort"`
	Username      string `xml:"username"`
	Password      string `xml:"password"`
	FromAddress   string `xml:"fromAddress"`
}

type KV struct {
	Key string `xml:"key,attr"`
	Val string `xml:",innerxml"`
}
type Api struct {
	Name     string `xml:"name,attr"`
	Url      string `xml:"url"`
	Method   string `xml:"method"`
	PostVals []KV   `xml:"postval"`
}

type ApiCall struct {
	Name string `xml:"name,attr"`
}

func (a *Api) MakeApiCall(title string, body string) {

	switch a.Method {
	case "POSTFORM":
		data := url.Values{}
		for _, v := range a.PostVals {
			switch v.Val {
			case "#title":
				v.Val = title
			case "#body":
				v.Val = body
			}
			data.Add(v.Key, v.Val)

		}
		_, err := http.PostForm(a.Url, data)
		if err != nil {
			log.Println(err)
		}

	default:
		log.Printf("Method %s for API isn't real\n", a.Method)
	}
}

type Config struct {
	Smtp    SmtpConfig `xml:"smtp"`
	Sites   []Site     `xml:"site"`
	Servers []Server   `xml:"server"`
	Apis    []Api      `xml:"api"`
}

type Email struct {
	Address string `xml:"address,attr"`
}

type Server struct {
	HostName string                 `xml:"hostname,attr"`
	Emails   []Email                `xml:"email"`
	Apis     []ApiCall              `xml:"api"`
	Disks    []checks.DiskCheckDisk `xml:"disk"`
}

func (s *Server) RunChecks(r *reporter.Reporter) {
	dc := &checks.DiskCheck{
		HostName:   s.HostName,
		CheckDisks: s.Disks,
	}
	err := dc.RunCheck(r)
	if err != nil {
		r.AddError(err)
	}
}

type Site struct {
	Name          string                  `xml:"name,attr"`
	Emails        []Email                 `xml:"email"`
	Apis          []ApiCall               `xml:"api"`
	RedirectTests []*checks.RedirectCheck `xml:"redirect"`
	SearchTests   []*checks.SearchCheck   `xml:"search"`
	CommandTests  []*checks.CommandCheck  `xml:"command"`
}

type Check interface {
	RunCheck(*reporter.Reporter) error
}

func flagWg(wg *sync.WaitGroup, donechan chan bool) {
	wg.Wait()
	donechan <- true
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

func DoApi(name string, subject string, body string) {
	var api *Api
	for _, thisApi := range config.Apis {
		if thisApi.Name == name {
			api = &thisApi
			break
		}
	}
	if api == nil {
		log.Printf("No api call named %s\n", name)
		return
	}
	api.MakeApiCall(subject, body)

}

func SendEmail(to string, subject string, body string) {
	headers := map[string]string{
		"To":      to,
		"From":    config.Smtp.FromAddress,
		"Subject": subject, //fmt.Sprintf("%s FAILED TEST", site.Name),
	}

	headerString := ""
	for k, v := range headers {
		headerString += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	auth := smtp.PlainAuth("", config.Smtp.Username, config.Smtp.Password, config.Smtp.ServerAddress)

	err := smtp.SendMail(config.Smtp.ServerAddress+":"+config.Smtp.ServerPort, auth, to, []string{to}, []byte(headerString+"\r\n"+body))
	if err != nil {
		fmt.Printf("Error Sending Mail: %s\n", err)
		fmt.Printf("%s:%s\n", config.Smtp.ServerAddress, config.Smtp.ServerPort)
		fmt.Printf("%s\n", headerString)
	}
}

var config Config

func DoWarnings(emails []Email, apis []ApiCall, r *reporter.Reporter) {
	r.DumpReport()
	warnings := r.CollectWarnings()
	if len(warnings) < 1 {
		return
	}

	subject := r.Name
	body := strings.Join(warnings, "\n")
	for _, email := range emails {
		SendEmail(email.Address, subject, body)
	}
	for _, api := range apis {
		DoApi(api.Name, subject, body)
	}

}

func main() {
	flag.Parse()
	file, err := os.Open(configFilename) //"/etc/informer/sites.xml")
	if err != nil {
		log.Panic(err)
	}

	decoder := xml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	for _, site := range config.Sites {
		r := reporter.GetRoot(site.Name)
		site.RunChecks(r)
		DoWarnings(site.Emails, site.Apis, r)
	}

	wg := sync.WaitGroup{}
	for _, server := range config.Servers {
		wg.Add(1)
		go func(server Server) {
			defer wg.Done()
			r := reporter.GetRoot(server.HostName)
			server.RunChecks(r)
			DoWarnings(server.Emails, server.Apis, r)
		}(server)
	}
	wg.Wait()
}
