package objects

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	ServerAddress string `xml:"serverAddress"`
	ServerPort    string `xml:"serverPort"`
	Username      string `xml:"username"`
	Password      string `xml:"password"`
	FromAddress   string `xml:"fromAddress"`
}

func (smtpConfig *Mailer) SendEmail(to string, subject string, body string) {
	headers := map[string]string{
		"To":      to,
		"From":    smtpConfig.FromAddress,
		"Subject": subject, //fmt.Sprintf("%s FAILED TEST", site.Name),
	}

	headerString := ""
	for k, v := range headers {
		headerString += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.ServerAddress)

	err := smtp.SendMail(smtpConfig.ServerAddress+":"+smtpConfig.ServerPort, auth, to, []string{to}, []byte(headerString+"\r\n"+body))
	if err != nil {
		fmt.Printf("Error Sending Mail: %s\n", err)
		fmt.Printf("%s:%s\n", smtpConfig.ServerAddress, smtpConfig.ServerPort)
		fmt.Printf("%s\n", headerString)
	}
}
