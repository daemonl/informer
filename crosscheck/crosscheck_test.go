package crosscheck

import (
	"sync"
	"testing"

	"github.com/daemonl/informer/reporter"
)

func TestCrosscheck(t *testing.T) {

	r := reporter.GetRoot("Test")

	ports := []string{"55550", "55551", "55553"}

	allRemotes := make([]string, len(ports), len(ports))

	for i, port := range ports {
		allRemotes[i] = "http://localhost:" + port
	}

	// Should all succeed, with one elected host
	anyGotElected := false
	w := sync.WaitGroup{}
	for _, port := range ports {
		w.Add(1)
		go func(port string) {
			elected := Crosscheck(&CXConfig{
				Bind:    ":" + port,
				Remotes: allRemotes,
			}, "asdf", r.Spawn("Child %s", port))
			if elected {
				anyGotElected = true
			}
			w.Done()
		}(port)
	}
	w.Wait()
	if !anyGotElected {
		t.Log("No host was elected")
		t.Fail()
	}
	r.DumpReport()

	allRemotes = append(allRemotes, "http://10.10.10.10:5000")
	w = sync.WaitGroup{}
	for _, port := range ports {
		w.Add(1)
		go func(port string) {
			elected := Crosscheck(&CXConfig{
				Bind:     ":" + port,
				Remotes:  allRemotes,
				MaxTries: 1,
			}, "asdf", r.Spawn("Child %s", port))
			if !elected {
				t.Log("Host not elected on fail")
				t.Fail()
			}
			w.Done()
		}(port)
	}
	w.Wait()
	r.DumpReport()
}
