package checks

import (
	"bufio"
	"strings"
	"time"

	"github.com/daemonl/informer/reporter"
)

type LogCheck struct {
	Request
	Format      string `xml:"format"`
	QuietPeriod string `xml:"quietPeriod"`
}

func (t *LogCheck) RunCheck(r *reporter.Reporter) error {

	resp, err := t.Request.DoRequest()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	lastTime := time.Unix(0, 0)

	for scanner.Scan() {
		str := scanner.Text()
		timeString := strings.Split(str, "-")[0]
		t, _ := time.Parse(t.Format, timeString)
		if t.After(lastTime) {
			lastTime = t
		}
	}

	if len(t.QuietPeriod) > 0 {
		res := r.Report("LOG CHECK %s", t.Url)
		duration, err := time.ParseDuration(t.QuietPeriod)
		if err != nil {
			return err
		}
		since := time.Since(lastTime)
		if since > duration {
			res.Fail("Last log was %s ago at %s", since, lastTime.Format(time.RFC1123))
		} else {
			res.Pass("Last log was %s ago", since)
		}
	}

	return nil

}
