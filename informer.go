package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/daemonl/informer/objects"

	"sync"
)

var configDir string
var runGroup string
var dryRun bool
var solo bool

func init() {
	flag.StringVar(&configDir, "config", "/etc/informer/conf.d/", "Config directory")
	flag.StringVar(&runGroup, "group", "", "When set, only run this group, otherwise runs root")
	flag.BoolVar(&dryRun, "dry", false, "When true, won't send mail or call APIs")
	flag.BoolVar(&solo, "solo", false, "Skip Crosschecks")
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
	s, err := dir.Stat()
	if err != nil {
		return nil, err
	}
	cfg := &objects.Core{}
	if !s.IsDir() {
		decoder := xml.NewDecoder(dir)
		err = decoder.Decode(&cfg)
		return cfg, err
	}

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
			return nil, fmt.Errorf("Decoding %s: %s", fInfo.Name(), err.Error())
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

	core.Run(runGroup)

}
