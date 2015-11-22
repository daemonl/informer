package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/daemonl/informer/crosscheck"
	"github.com/daemonl/informer/objects"
	"github.com/daemonl/informer/reporter"

	"sync"
)

var configDir string
var runGroup string
var dryRun bool

func init() {
	flag.StringVar(&configDir, "config", "/etc/informer/conf.d/", "Config directory")
	flag.StringVar(&runGroup, "group", "", "When set, only run this group, otherwise runs root")
	flag.BoolVar(&dryRun, "dry", false, "When true, won't send mail or call APIs")
}

func flagWg(wg *sync.WaitGroup, donechan chan bool) {
	wg.Wait()
	donechan <- true
}

type FilesByName []os.FileInfo

func (s FilesByName) Len() int           { return len(s) }
func (s FilesByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s FilesByName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }

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
	sort.Sort(FilesByName(files))
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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}

	if core.Crosscheck != nil {

		configHash, err := crosscheck.XmlHash(core)
		if err != nil {
			log.Printf("Couldn't encode config: %s\n", err.Error())
		}
		log.Printf("Config %s\n", configHash)
		r := reporter.GetRoot("Crosscheck")
		elected := crosscheck.Crosscheck(core.Crosscheck, configHash, r)
		r.DumpReport()
		if !dryRun {
			core.DoWarnings(r, &core.Admins)
		}
		if !elected {
			return
		}
	}

	list := map[string][]objects.Group{}

	for _, group := range core.Groups {
		// Matches "", which is 'unspecified'
		if runGroup == "all" || runGroup == group.RunGroup {
			_, ok := list[group.SyncGroup]
			if !ok {
				list[group.SyncGroup] = []objects.Group{}
			}
			list[group.SyncGroup] = append(list[group.SyncGroup], group)
		}
	}

	times := map[string]int64{}
	wg := sync.WaitGroup{}
	for name, sg := range list {
		//fmt.Printf("Run sync %s - %d groups\n", name, len(sg))
		wg.Add(1)
		go func(name string, sg []objects.Group) {
			defer wg.Done()
			start := time.Now().Unix()
			defer func() { times[name] = time.Now().Unix() - start }()
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
				if !dryRun {
					core.DoWarnings(r, &group.Informants)
				}
			}

		}(name, sg)
	}
	wg.Wait()

	for name, seconds := range times {
		fmt.Printf("%s took %d seconds\n", name, seconds)
	}

}
