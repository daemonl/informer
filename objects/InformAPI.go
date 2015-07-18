package objects

import (
	"log"
	"net/http"
	"net/url"
)

type InformAPI struct {
	Name     string `xml:"name,attr"`
	Url      string `xml:"url"`
	Method   string `xml:"method"`
	PostVals []struct {
		Key string `xml:"key,attr"`
		Val string `xml:",innerxml"`
	} `xml:"postval"`
}

func (a *InformAPI) Call(title string, body string) {

	switch a.Method {
	case "POSTFORM":
		data := url.Values{}
		for _, v := range a.PostVals {
			switch v.Val {
			case "#title":
				v.Val = title
			case "#body":
				v.Val = body
			}
			data.Add(v.Key, v.Val)

		}
		_, err := http.PostForm(a.Url, data)
		if err != nil {
			log.Println(err)
		}

	default:
		log.Printf("Method %s for API isn't real\n", a.Method)
	}
}
