package objects

type Core struct {
	Mailer *Mailer `xml:"smtp"`
	Checks
	Groups []Group `xml:"group"`
}

type Group struct {
	Checks
	Name string `xml:"name,attr"`
}

type Checks struct {
	Sites   []Site        `xml:"site"`
	Servers []ServerCheck `xml:"server"`
	Apis    []InformAPI   `xml:"api"`
	Data    []DataCheck   `xml:"data"`
}
