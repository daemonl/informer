package checks

import (
	"crypto/tls"
	"fmt"
	"time"
	//	"log"
	"net/http"
)

type CustomClient struct {
	CertFile *string `xml:"cert"`
	KeyFile  *string `xml:"key"`
	Insecure bool    `xml:"insecure,attr"`
	Timeout  int64   `xml:"timeout,attr"`
}

func (c *CustomClient) HashBase() string {
	if c == nil {
		return ""
	}
	s := fmt.Sprintf("%b %d", c.Insecure, c.Timeout)
	if c.CertFile != nil {
		s += " " + *c.CertFile
	}
	if c.KeyFile != nil {
		s += " " + *c.KeyFile
	}
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

	if c.CertFile != nil || c.KeyFile != nil {

		if c.CertFile == nil || c.KeyFile == nil {
			return nil, fmt.Errorf("Either both or neither key and cert must be set for custom clients")
		}
		clientCertificate, err := tls.LoadX509KeyPair(*c.CertFile, *c.KeyFile)
		if err != nil {
			return nil, err
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.Certificates = []tls.Certificate{
			clientCertificate,
		}
	}

	if c.Insecure {
		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.InsecureSkipVerify = true
	}
	return &http.Client{Transport: tr}, nil

}
