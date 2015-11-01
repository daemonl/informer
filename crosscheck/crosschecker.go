package crosscheck

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/daemonl/informer/reporter"
)

type CXConfig struct {
	Bind     string   `xml:"bind"`
	Remotes  []string `xml:"remote"`
	MaxTries int      `xml:"max"`
}

func Crosscheck(cfg *CXConfig, r *reporter.Reporter) bool {

	if cfg.MaxTries == 0 {
		cfg.MaxTries = 3
	}

	b := make([]byte, 21, 21)
	rand.Read(b)
	myID := base64.StdEncoding.EncodeToString(b)
	log.Printf("My ID: %s\n", myID)

	mux := http.NewServeMux()
	// Serve and dial
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(myID))
	})
	go http.ListenAndServe(cfg.Bind, mux)

	isLowest := true
	allSuccess := true
	for _, remote := range cfg.Remotes {
		reportRes := r.Report("CX %s", remote)
		success := false
		for try := 0; try < cfg.MaxTries; try += 1 {
			time.Sleep(time.Second * time.Duration(1+(5*try)))
			res, err := http.Get(remote)
			if err != nil {
				log.Printf("Checking %s fail %d/%d\n%s", remote, try+1, cfg.MaxTries, err.Error())
				continue
			}
			bodyBytes, _ := ioutil.ReadAll(res.Body)
			remoteID := string(bodyBytes)
			reportRes.Pass("Got ID %s", remoteID)
			success = true
			if remoteID == myID {
				log.Printf("Self Check OK %s\n", remoteID)
				break
			}
			log.Printf("Remote %s answered %s\n", remote, remoteID)
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

	if isLowest {
		log.Printf("I am the lowest: %s\n", myID)
		return true
	}
	log.Println("Not elected, I'm done")
	return false
}
