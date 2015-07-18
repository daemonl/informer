package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"

	"github.com/daemonl/informer/objects"
	"github.com/daemonl/informer/reporter"

	"sync"
)

var configDir string
var runGroup string

func init() {
	flag.StringVar(&configDir, "config", "/etc/informer/conf.d/", "Config directory")
	flag.StringVar(&runGroup, "group", "", "When set, only run this group, otherwise runs root")
}

func flagWg(wg *sync.WaitGroup, donechan chan bool) {
	wg.Wait()
	donechan <- true
}

func loadConfig(dirName string) (*objects.Core, error) {
	dir, err := os.Open(os.ExpandEnv(dirName))
	if err != nil {
		return nil, err
	}
	cfg := &objects.Core{}

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, fInfo := range files {
		file, err := os.Open(dirName + fInfo.Name())
		if err != nil {
			return nil, err
		}
		decoder := xml.NewDecoder(file)
		err = decoder.Decode(&cfg)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func main() {
	flag.Parse()

	core, err := loadConfig(configDir)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
		return
	}

	list := []objects.Checks{}

	if runGroup == "" {
		list = append(list, core.Checks)
	} else {
		for _, group := range core.Groups {
			if runGroup == "all" || runGroup == group.Name {
				list = append(list, group.Checks)
			}
		}
	}

	wg := sync.WaitGroup{}
	for _, group := range list {
		for _, server := range group.Servers {
			wg.Add(1)
			go func(server objects.ServerCheck) {
				defer wg.Done()
				r := reporter.GetRoot(server.Name)
				server.RunChecks(r)
				server.DoWarnings(core, r)
			}(server)
		}
	}
	wg.Wait()

	for _, group := range list {
		for _, ds := range group.Data {
			r := reporter.GetRoot(ds.Name)
			ds.RunChecks(r)
			ds.DoWarnings(core, r)
		}
		for _, site := range group.Sites {
			r := reporter.GetRoot(site.Name)
			site.RunChecks(r)
			site.DoWarnings(core, r)
		}
	}
}
