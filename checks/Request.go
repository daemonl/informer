package checks

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Pairs []Pair

func (pairs Pairs) Values() url.Values {
	v := url.Values{}
	for _, entry := range pairs {
		v.Add(entry.Name, entry.Value)
	}
	return v
}

type Pair struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

type Request struct {
	Url          string        `xml:"url,attr"`
	CustomClient *CustomClient `xml:"client"`
	Method       string        `xml:"method,attr"`
	FormVals     Pairs         `xml:"form"`
	Body         string        `xml:"body"`
	Cookies      Pairs         `xml:"cookie"`
	Headers      Pairs         `xml:"header"`
}

func (r *Request) DoRequest() (*http.Response, error) {

	client, err := r.CustomClient.GetClient()
	if err != nil {
		return nil, err
	}
	if r.Method == "" {
		if len(r.FormVals) > 0 {
			r.Method = "POST"
		} else {
			r.Method = "GET"
		}
	}

	var body io.Reader

	headers := r.Headers.Values()

	if len(r.FormVals) > 0 {
		headers.Add("Content-Type", "application/x-www-form-urlencoded")
		body = strings.NewReader(r.FormVals.Values().Encode())
	} else if r.Method == "POST" || r.Method == "PUT" {
		body = strings.NewReader(r.Body)
	}

	req, err := http.NewRequest(r.Method, r.Url, body)
	if err != nil {
		return nil, err
	}
	for key, vals := range headers {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}
	for _, cookie := range r.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: strings.TrimSpace(cookie.Value),
		})
	}
	return client.Do(req)
}
