package checks

import (
	"crypto/tls"
	"fmt"
	//	"log"
	"net/http"
)

type CustomClient struct {
	CertFile *string `xml:"cert"`
	KeyFile  *string `xml:"key"`
}

func (c *CustomClient) GetClient() (*http.Client, error) {
	if c == nil {
		client := &http.Client{}
		return client, nil
	}
	tr := &http.Transport{}

	if c.CertFile != nil || c.KeyFile != nil {

		if c.CertFile == nil || c.KeyFile == nil {
			return nil, fmt.Errorf("Either both or neither key and cert must be set for custom clients")
		}
		//log.Printf("Using %s and %s for client auth\n", *c.CertFile, *c.KeyFile)
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

	return &http.Client{Transport: tr}, nil

}
