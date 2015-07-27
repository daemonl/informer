package checks

import (
	"io/ioutil"
	"strings"

	"github.com/daemonl/informer/reporter"
)

type TextCheck struct {
	Request
	Contains []string `xml:"contains"`
}

func (t *TextCheck) RunCheck(r *reporter.Reporter) error {

	reader, err := t.GetReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	bodyBytes, err := ioutil.ReadAll(reader)
	bodyString := string(bodyBytes)

	for _, ss := range t.Contains {
		res := r.Report("SEARCH %s for %s", t.URL, ss)
		if !strings.Contains(bodyString, ss) {
			res.Fail("Did not contain '%s'\n\n%s", ss, bodyString)
			return nil
		} else {
			res.Pass("Found")
		}
	}
	return nil
}
