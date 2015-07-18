package checks

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/daemonl/informer/reporter"
)

type SearchCheck struct {
	Request
	Contains []string `xml:"string"`
}

func (t *SearchCheck) RunCheck(r *reporter.Reporter) error {

	resp, err := t.Request.DoRequest()
	if err != nil {
		switch err := err.(type) {
		case *url.Error:

			switch err := err.Err.(type) {
			case *net.OpError:

				switch err := err.Err.(type) {
				case *net.DNSError:
					return fmt.Errorf("DNS Lookup Error: %s", err.Err)
				default:
					return err
				}

			default:
				return err
			}

		default:
			return err
		}

	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	for _, ss := range t.Contains {
		res := r.Report("SEARCH %s for %s", t.Url, ss)
		if !strings.Contains(bodyString, ss) {
			res.Fail("Did not contain '%s'\n\n%s", ss, bodyString)
			return nil
		} else {
			res.Pass("Found")
		}
	}
	return nil
}
