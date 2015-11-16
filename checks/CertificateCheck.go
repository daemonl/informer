package checks

import (
	"crypto/tls"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/daemonl/informer/reporter"
)

type CertificateCheck struct {
	URL  string `xml:"url,attr"`
	Days uint64 `xml:"days"`
}

func (t *CertificateCheck) GetHash() string {
	return hashFromf("CERT:%s %s", t.URL, t.Days)
}

func (t *CertificateCheck) GetName() string {
	return t.URL
}

func (t *CertificateCheck) RunCheck(r *reporter.Reporter) error {
	u, err := url.Parse(t.URL)
	if err != nil {
		return err
	}
	port := "443"
	host := u.Host
	if strings.Contains(host, ":") {
		parts := strings.Split(u.Host, ":")
		host = parts[0]
		port = parts[1]
	}
	conn, err := tls.Dial("tcp", host+":"+port, nil)
	if err != nil {
		return err
	}

	if t.Days == 0 {
		t.Days = 15
	}
	res := r.Report("Certificate expires > %d days", t.Days)
	cert := conn.ConnectionState().PeerCertificates[0]
	conn.Close()
	exp := time.Since(cert.NotAfter) * -1
	days := uint64(math.Floor(exp.Hours() / 24))
	if days < t.Days {
		res.Fail("Certificate expires in %d days (< %d)", days, t.Days)
	} else {
		res.Pass("Certificate expires in %d days", days)
	}

	return nil
}
