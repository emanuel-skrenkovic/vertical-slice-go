package core

import (
	"fmt"
	"net/smtp"
	"net/url"
	"strings"
)

const htmlMime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"

type MailMessage struct {
	Subject    string
	From       string
	To         []string
	Cc         []string
	Bcc        []string
	BodyString string
	IsHTML     bool
}

func (m MailMessage) Content() []byte {
	subject := fmt.Sprintf("Subject: %s\n", m.Subject)
	mime := "\n"
	if m.IsHTML {
		mime = htmlMime
	}

	cc := ""
	if len(m.Cc) > 0 {
		cc = fmt.Sprintf("Cc: %s\n", strings.Join(m.Cc, ","))
	}

	bcc := ""
	if len(m.Bcc) > 0 {
		bcc = fmt.Sprintf("Bcc: %s\n", strings.Join(m.Bcc, ","))
	}

	return []byte(subject + mime + cc + bcc + "\n" + m.BodyString)
}

type EmailClient struct {
	host string
	auth smtp.Auth
}

func NewEmailClient(
	host *url.URL,
	username string,
	password string,
) (*EmailClient, error) {
	authHost := host.Host

	parts := strings.Split(host.Host, ":")
	if len(parts) > 1 {
		authHost = parts[0]
	}

	return &EmailClient{
		auth: smtp.PlainAuth("", username, password, authHost),
		host: host.Host,
	}, nil
}

func (c *EmailClient) Send(m MailMessage) error {
	return smtp.SendMail(c.host, c.auth, m.From, m.To, m.Content())
}
