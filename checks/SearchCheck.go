package checks

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/daemonl/informer/reporter"
)

type SearchCheck struct {
	Url          string        `xml:"url,attr"`
	Insecure     bool          `xml:"insecure,attr"`
	Contains     []string      `xml:"string"`
	CustomClient *CustomClient `xml:"client"`
	Cookies      []struct {
		Name  string `xml:"name,attr"`
		Value string `xml:",innerxml"`
	} `xml:"cookie"`
}

func (t *SearchCheck) RunCheck(r *reporter.Reporter) error {

	client, err := t.CustomClient.GetClient()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", t.Url, nil)
	if err != nil {
		return err
	}
	for _, cookie := range t.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: strings.TrimSpace(cookie.Value),
		})
	}
	if t.Insecure && strings.HasPrefix(t.Url, "https://") {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	resp, err := client.Do(req)
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
