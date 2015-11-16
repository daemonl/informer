package crosscheck

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/daemonl/informer/reporter"
)

type CXConfig struct {
	Bind     string   `xml:"bind"`
	Remotes  []string `xml:"remote"`
	MaxTries int      `xml:"max"`
}

func XmlHash(o interface{}) (string, error) {

	h := md5.New()
	err := xml.NewEncoder(h).Encode(o)
	if err != nil {
		return "", err
	}
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return hash, nil
}

type CrosscheckResult struct {
	AllResponded bool
	Total        int
	Offset       int
}

type InformerInstance struct {
	Address  string
	Hash     string
	ConnTo   net.Conn
	ConnFrom net.Conn
	Error    error
}

func (i *InformerInstance) Connect(maxTries int) {

	for try := 0; try < maxTries; try += 1 {
		time.Sleep(time.Second * time.Duration(1+(5*try)))
		i.ConnTo, i.Error = net.DialTimeout("tcp", i.Address, time.Second*3)
		if i.Error == nil {
			break
		}

		scanner := bufio.NewScanner(i.ConnTo)
		if !scanner.Scan() {
			err := fmt.Errorf("Couldn't scan line on connection")
			i.ConnTo.Close()
			break
		}

	}
	return
	/*
			bodyBytes, _ := ioutil.ReadAll(res.Body)
			parts := strings.Split(string(bodyBytes), "\n")
			if len(parts) != 2 {
				log.Println(string(bodyBytes))
				reportRes.Fail("Didn't understand response")
				cxResult.AllResponded = false
				continue remotes

			}
			remoteID := parts[0]
			remoteConfig := parts[1]
			reportRes.Pass("Got ID %s", remoteID)
			if remoteConfig != configHash {
				reportRes.Fail("Config mismach")
				cxResult.AllResponded = false
				continue remotes
			}
			success = true
			if remoteID == myID {
				log.Printf("Remote %s is me (%s)\n", remote, remoteID)
				break
			}
			log.Printf("Remote %s answered (%s)\n", remote, remoteID)
			allHashes = append(allHashes, remoteID)
			break
		}
	*/
}

func Crosscheck(cfg *CXConfig, configHash string, r *reporter.Reporter) (CrosscheckResult, error) {

	cxResult := CrosscheckResult{
		AllResponded: true,
		Total:        len(cfg.Remotes),
	}

	if cfg.MaxTries == 0 {
		cfg.MaxTries = 3
	}

	myID := uuid.New()
	log.Printf("My ID: %s\n", myID)

	l, err := net.Listen("tcp", cfg.Bind)
	if err != nil {
		cxResult.AllResponded = false
		return cxResult, err
	}

	connectWait := &sync.WaitGroup{}
	allInstances := map[string]*InformerInstance{}
	for _, remote := range cfg.Remotes {
		i := &InformerInstance{
			Address: remote,
		}
		allInstances[remote] = i
		connectWait.Add(1)
		go func() {
			i.Connect(cfg.MaxTries)
			connectWait.Done()
		}()
	}

	connectWait.Wait()

	for _, instance := range allInstances {
		if instance.Error != nil {
			cxResult.AllResponded = false
			return cxResult, err
		}
	}

	allHashes := make([]string, 0, len(cfg.Remotes))
remotes:
	for _, remote := range cfg.Remotes {
		reportRes := r.Report("CX %s", remote)
		success := false
		if !success {
			reportRes.Fail("No connection established")
			cxResult.AllResponded = false
		}
	}
	if !cxResult.AllResponded {
		log.Printf("At least one host failed, running checks locally")
		return cxResult
	}

	sort.Strings(allHashes)

	myIndex := -1
	for i, h := range allHashes {
		if h == myID {
			myIndex = i
		}
	}
	if myIndex < 0 {
		cxResult.AllResponded = false
		return cxResult
	}

	cxResult.Offset = myIndex

	return cxResult
}
