package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type Sender struct {
	host     string
	port     string
	user     string
	password string
}

func NewSender(host, port, user, password string) *Sender {
	return &Sender{host: host, port: port, user: user, password: password}
}

// Send sends an HTML email via SMTP (STARTTLS on port 587)
func (s *Sender) Send(to, subject, htmlBody string) error {
	addr := s.host + ":" + s.port

	msg := []byte(fmt.Sprintf(
		"From: Finbud <hello.finbudapp@gmail.com>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s",
		to, subject, htmlBody,
	))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp baglanti hatasi: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client hatasi: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{ServerName: s.host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls hatasi: %w", err)
	}

	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth hatasi: %w", err)
	}

	if err := client.Mail(s.user); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}
