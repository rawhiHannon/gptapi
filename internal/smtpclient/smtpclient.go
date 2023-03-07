package smtpclient

import (
	"gptapi/pkg/enum"
	"log"
	"strings"

	"github.com/wneessen/go-mail"
)

func getServerTypeOption(serverType enum.SMTPType) mail.Option {
	if serverType == enum.TLS {
		return mail.WithTLSPolicy(mail.TLSMandatory)
	}
	return mail.WithSSL()
}

type SMTPClient struct {
	smtpType enum.SMTPType
	host     string
	port     int
	from     string
	username string
	password string
}

func New(smtpType enum.SMTPType, host string, port int, from, username, password string) *SMTPClient {
	obj := &SMTPClient{
		smtpType: smtpType,
		host:     host,
		port:     port,
		from:     from,
		username: username,
		password: password,
	}
	return obj
}

func (s *SMTPClient) getClient() *mail.Client {
	c, err := mail.NewClient(
		s.host,
		mail.WithPort(s.port),
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
		getServerTypeOption(s.smtpType),
	)
	if err != nil {
		log.Println("failed to create mail client:", err)
	}
	return c
}

func (s *SMTPClient) sendMail(tomail string, subject string, msg string) error {
	m := mail.NewMsg()
	if err := m.From(s.from); err != nil {
		log.Fatalf("Invalid `From` address: %s", err)
	}
	to := strings.Split(tomail, ",")
	if err := m.To(to...); err != nil {
		log.Fatalf("Invalid `To` address: %s", err)
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, msg)
	client := s.getClient()
	if err := client.DialAndSend(m); err != nil {
		return err
	}
	log.Println("Email sent successfully to:", tomail)
	return nil
}

func (s *SMTPClient) Send(tomail string, subject string, msg string) {
	var err error
	err = s.sendMail(tomail, subject, msg)
	if err != nil {
		log.Fatalf("Failed to send mail: %s", err)
	}
}
