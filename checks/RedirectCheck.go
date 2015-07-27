package checks

import (
	"github.com/daemonl/informer/reporter"
	"net/http"
)

type RedirectCheck struct {
	From         string        `xml:"from"`
	To           string        `xml:"to"`
	CustomClient *CustomClient `xml:"client"`
}

func (t *RedirectCheck) GetName() string {
	return t.From
}

func (t *RedirectCheck) RunCheck(r *reporter.Reporter) error {
	res := r.Report("CHECK REDIRECT %s => %s", t.From, t.To)
	client := &http.Client{}
	redirected := false
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if redirected {
			res.Fail("Redirected to subsequent address %s", req.URL.String())
		}
		redirected = true

		if req.URL.String() != t.To {
			res.Fail("Redirect to wrong address %s", req.URL.String())
		}
		return nil
	}
	resp, err := client.Get(t.From)
	if err != nil {
		res.Fail("Redirect %s => %s failed:\n %s", t.From, t.To, err.Error())
		return nil
	}
	if !redirected {
		res.Fail("Redirect %s => %s did not redirect", t.From, t.To)
	}
	resp.Body.Close()
	return nil
}
