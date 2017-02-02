package checks

import (
	"crypto/tls"
	"fmt"
	"time"
	//	"log"
	"net/http"
)

type CustomClient struct {
	Cert     *[]byte `xml:"cert"`
	Key      *[]byte `xml:"key"`
	Insecure bool    `xml:"insecure,attr"`
	Timeout  int64   `xml:"timeout,attr"`
}

func (c *CustomClient) HashBase() string {
	if c == nil {
		return ""
	}
	s := fmt.Sprintf("%b %d", c.Insecure, c.Timeout)
	return s
}

func (c *CustomClient) GetClient() (*http.Client, error) {
	if c == nil {
		client := &http.Client{}
		return client, nil
	}
	tr := &http.Transport{}

	if c.Timeout == 0 {
		c.Timeout = 10
	}
	tr.ResponseHeaderTimeout = time.Duration(c.Timeout) * time.Second

	if c.Cert != nil || c.Key != nil {

		if c.Cert == nil || c.Key == nil {
			return nil, fmt.Errorf("Either both or neither key and cert must be set for custom clients")
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}

		cert, err := tls.X509KeyPair(*c.Cert, *c.Key)
		if err != nil {
			return nil, err
		}
		tr.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	if c.Insecure {
		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.InsecureSkipVerify = true
	}
	return &http.Client{Transport: tr}, nil

}
