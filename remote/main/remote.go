package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/daemonl/informer/remote/commands"
)

var configFilename string

func init() {
	flag.StringVar(&configFilename, "config", "config.json", "The config file")
}

type Config struct {
	ServerCertificate string `xml:"serverCertificate" json:"serverCertificate"`
	ServerPrivate     string `xml:"serverPrivate" json:"serverPrivate"`
	BindAddress       string `xml:"bindAddress" json:"bindAddress"`
}

func loadConfig(configFilename string) (*Config, error) {
	config := &Config{}
	configFile, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	type iDecoder interface {
		Decode(interface{}) error
	}
	var dec iDecoder

	if strings.HasSuffix(configFilename, ".xml") {
		dec = xml.NewDecoder(configFile)
	} else if strings.HasSuffix(configFilename, ".json") {
		dec = json.NewDecoder(configFile)
	} else {
		return nil, fmt.Errorf("Config filetype not supporter")
	}

	err = dec.Decode(config)
	return config, err
}

func main() {
	flag.Parse()
	config, err := loadConfig(configFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
		return
	}

	err = serve(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type CommandRunner interface {
	GetCmd() *exec.Cmd
	ParseOutput(io.Reader) (interface{}, error)
}

type Command struct {
	CommandRunner
}

func (c *Command) Run(w io.Writer) error {
	cmd := c.GetCmd()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	/*
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}*/

	cmd.Start()
	obj, err := c.ParseOutput(stdout)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}

func (c Command) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := c.Run(w)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func serve(config *Config) error {
	http.Handle("/df", &Command{CommandRunner: &commands.DiskCommand{}})

	certFile := os.ExpandEnv(config.ServerCertificate)
	keyFile := os.ExpandEnv(config.ServerPrivate)
	caCertFile := os.ExpandEnv(config.ServerCertificate)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    x509.NewCertPool(),
	}
	caCerts, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return err
	}
	tlsConfig.ClientCAs.AppendCertsFromPEM(caCerts)
	l, err := tls.Listen("tcp", config.BindAddress, tlsConfig)
	if err != nil {
		return err
	}

	return http.Serve(l, nil)
}
