package crosscheck

import (
	"sync"
	"testing"

	"github.com/daemonl/informer/reporter"
)

func TestCrosscheck(t *testing.T) {

	r := reporter.GetRoot("Test")

	ports := []string{"55550"} //, "55551", "55553"}

	allRemotes := make([]string, len(ports), len(ports))

	for i, port := range ports {
		allRemotes[i] = "http://localhost:" + port
	}

	// Should all succeed, with one elected host
	w := sync.WaitGroup{}
	results := make([]CrosscheckResult, len(ports), len(ports))
	for i, port := range ports {
		w.Add(1)
		go func(port string) {
			cxResult := Crosscheck(&CXConfig{
				Bind:    ":" + port,
				Remotes: allRemotes,
			}, "H", r.Spawn("Child %s", port))
			results[i] = cxResult
			if cxResult.Total != len(ports) {
				t.Logf("Wrong total %d", cxResult.Total)
				t.Fail()
			}
			w.Done()

		}(port)
	}
	w.Wait()

	offsetExists := make([]bool, len(ports), len(ports))
	for _, r := range results {
		offsetExists[r.Offset] = true
	}

	for i, b := range offsetExists {
		if !b {
			t.Logf("Missing offset %d", i)
			t.Fail()
		}
	}
	r.DumpReport()

	allRemotes = append(allRemotes, "http://10.10.10.10:5000")
	w = sync.WaitGroup{}
	for _, port := range ports {
		w.Add(1)
		go func(port string) {
			cxResult := Crosscheck(&CXConfig{
				Bind:     ":" + port,
				Remotes:  allRemotes,
				MaxTries: 1,
			}, "H", r.Spawn("Child %s", port))
			if cxResult.AllResponded {
				t.Log("Host not elected on fail")
				t.Fail()
			}
			w.Done()
		}(port)
	}
	w.Wait()
	r.DumpReport()
}
