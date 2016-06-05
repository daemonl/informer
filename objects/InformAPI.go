package objects

import (
	"log"
	"net/http"
	"net/url"
)

type InformAPI struct {
	Name     string `xml:"name,attr"`
	Url      string `xml:"url,omitempty"`
	Method   string `xml:"method,omitempty"`
	PostVals []struct {
		Key string `xml:"key,attr"`
		Val string `xml:",innerxml"`
	} `xml:"postval,omitempty"`
}

type InformParams map[string]string

func (a *InformAPI) Call(p InformParams) {

	switch a.Method {
	case "POSTFORM":
		data := url.Values{}
		for _, postVal := range a.PostVals {
			for replaceKey, replaceVal := range p {
				if postVal.Val == "#"+replaceKey {
					postVal.Val = replaceVal
				}
			}
			data.Add(postVal.Key, postVal.Val)
		}
		_, err := http.PostForm(a.Url, data)
		if err != nil {
			log.Println(err)
		}

	default:
		log.Printf("Method %s for API isn't real\n", a.Method)
	}
}
