package checks

import (
	"bufio"
	"regexp"
	"time"

	"github.com/daemonl/informer/reporter"
)

type LogCheck struct {
	Request
	Format      string `xml:"format"`
	QuietPeriod string `xml:"quietPeriod"`
	Regexp      string `xml:"regex"`
}

func (t *LogCheck) GetHash() string {
	return hashFromf("LOG:%s %s %s %s",
		t.Format,
		t.QuietPeriod,
		t.Regexp,
		t.Request.HashBase(),
	)
}

func (t *LogCheck) RunCheck(r *reporter.Reporter) error {

	resp, err := t.Request.DoRequest()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	reTimePart, err := regexp.Compile(t.Regexp)
	if err != nil {
		return err
	}

	lastTime := time.Unix(0, 0)

	for scanner.Scan() {
		str := scanner.Text()
		timeString := reTimePart.FindString(str)
		t, _ := time.Parse(t.Format, timeString)
		if t.After(lastTime) {
			lastTime = t
		}
	}

	if len(t.QuietPeriod) > 0 {
		res := r.Report("LOG CHECK %s", t.GetName())
		duration, err := time.ParseDuration(t.QuietPeriod)
		if err != nil {
			return err
		}
		since := time.Since(lastTime)
		if since > duration {
			res.Fail("FAIL Last log was %s ago at %s", since, lastTime.Format(time.RFC1123))
		} else {
			res.Pass("Last log was %s ago", since)
		}
	}

	return nil

}
