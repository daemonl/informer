package crosscheck

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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

func Crosscheck(cfg *CXConfig, configHash string, r *reporter.Reporter) bool {

	if cfg.MaxTries == 0 {
		cfg.MaxTries = 3
	}

	b := make([]byte, 21, 21)
	rand.Read(b)
	myID := base64.URLEncoding.EncodeToString(b)
	log.Printf("My ID: %s\n", myID)

	waitForOthers := &sync.WaitGroup{}
	waitForOthers.Add(len(cfg.Remotes))

	mux := http.NewServeMux()
	// Serve and dial
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		log.Printf("Was pinged by %s\n", id)
		waitForOthers.Done()
		fmt.Fprintf(w, "%s\n%s", myID, configHash)

	})

	go func() {
		err := http.ListenAndServe(cfg.Bind, mux)
		if err != nil {
			log.Printf("Can't serve for CX: %s\n", err.Error())
		}
	}()

	isLowest := true
	allSuccess := true
remotes:
	for _, remote := range cfg.Remotes {
		reportRes := r.Report("CX %s", remote)
		success := false
		for try := 0; try < cfg.MaxTries; try += 1 {
			time.Sleep(time.Second * time.Duration(1+(5*try)))
			res, err := http.Get(fmt.Sprintf("%s?id=%s", remote, myID))
			if err != nil {
				log.Printf("Checking %s fail %d/%d\n%s", remote, try+1, cfg.MaxTries, err.Error())
				continue
			}
			bodyBytes, _ := ioutil.ReadAll(res.Body)
			parts := strings.Split(string(bodyBytes), "\n")
			if len(parts) != 2 {
				log.Println(string(bodyBytes))
				reportRes.Fail("Didn't understand response")
				allSuccess = false
				continue remotes

			}
			remoteID := parts[0]
			remoteConfig := parts[1]
			reportRes.Pass("Got ID %s", remoteID)
			if remoteConfig != configHash {
				reportRes.Fail("Config mismach")
				allSuccess = false
				continue remotes
			}
			success = true
			if remoteID == myID {
				log.Printf("Remote %s is me (%s)\n", remote, remoteID)
				break
			}
			log.Printf("Remote %s answered (%s)\n", remote, remoteID)
			if remoteID < myID {
				isLowest = false
			}
			break
		}
		if !success {
			reportRes.Fail("No connection established")
			allSuccess = false
		}
	}
	if !allSuccess {
		log.Printf("At least one host failed, running checks locally")
		return true
	}

	log.Println("Waiting for other servers")
	waitForOthers.Wait()
	log.Println("Wait complete")

	if isLowest {
		log.Printf("I am the lowest: %s\n", myID)
		return true
	}
	log.Println("Not elected, I'm done")
	return false
}
