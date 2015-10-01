package service

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Services map[string]Service

type Service struct {
	Name    string `xml:"name,attr"`
	Status  bool   `xml:"status,attr"`
	Restart bool   `xml:"restart,attr"`
	Report  bool   `xml:"report,attr"`
}

func (s *Services) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	arr := []Service{}
	*s = map[string]Service{}
	err := d.DecodeElement(&arr, &start)
	if err != nil {
		return err
	}
	for _, service := range arr {
		(*s)[service.Name] = service
	}
	return nil
}

func (s Services) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.NotFound(w, r)
		return
	}
	serviceName := parts[2]
	action := parts[3]

	service := s[serviceName]
	switch action {
	case "status":
		if !service.Status {
			http.Error(w, "status not allowed", http.StatusForbidden)
			return
		}
	case "restart":
		if !service.Restart {
			http.Error(w, "restart not allowed", http.StatusForbidden)
			return
		}
	default:
		http.Error(w, action+" not recognized", http.StatusNotFound)
		return
	}
	log.Printf("%s %s\n", service.Name, action)

	var cmd *exec.Cmd
	if os.Getuid() == 0 {
		cmd = exec.Command("service",
			service.Name,
			action)
	} else {
		cmd = exec.Command("sudo",
			"service",
			service.Name,
			action)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Command error", http.StatusInternalServerError)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		http.Error(w, "Command error", http.StatusInternalServerError)
		return
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	go func() {
		io.Copy(w, stdout)
	}()
	err = cmd.Start()
	if err != nil {
		http.Error(w, "Command error", http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	err = cmd.Wait()
	if err != nil {
		http.Error(w, "Command error", http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	if action == "restart" {
		fmt.Fprintf(w, "Service %s restarted\n", service.Name)
	}
}
