package checks

import (
	"github.com/daemonl/informer/reporter"
	"net/http"
)

type RedirectCheck struct {
	From          string        `xml:"from"`
	To            string        `xml:"to"`
	CustomClient  *CustomClient `xml:"client"`
	AllowMultiple bool          `xml:"allow-multiple,attr"`
}

func (t *RedirectCheck) GetHash() string {
	return hashFromf("REDIRECT:%s>%s %s",
		t.From,
		t.To,
		t.CustomClient.HashBase(),
	)
}
func (t *RedirectCheck) GetName() string {
	return t.From
}

func (t *RedirectCheck) RunCheck(r *reporter.Reporter) error {
	res := r.Report("CHECK REDIRECT %s => %s", t.From, t.To)
	client := &http.Client{}
	finalAddress := ""
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !t.AllowMultiple {
			if finalAddress != "" {
				res.Fail("Redirected to subsequent address %s", req.URL.String())
			}
		}
		finalAddress = req.URL.String()
		return nil
	}

	resp, err := client.Get(t.From)
	if finalAddress != t.To {
		res.Fail("Redirect to wrong address %s", finalAddress)
	}
	if err != nil {
		res.Fail("Redirect %s => %s failed:\n %s", t.From, t.To, err.Error())
		return nil
	}
	if finalAddress == "" {
		res.Fail("Redirect %s => %s did not redirect", t.From, t.To)
	}
	resp.Body.Close()
	return nil
}
