package mailer

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type (
	Config struct {
		Host         string
		Port         int
		AuthEmail    string
		AuthPassword string
	}

	Mailer struct {
		emailConfig *Config
		Body        string
		Error       error
	}
)

func NewMailer() Mailer {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		panic("invalid smtp port")
	}

	emailConfig := &Config{
		Host:         os.Getenv("SMTP_HOST"),
		Port:         port,
		AuthEmail:    os.Getenv("SMTP_AUTH_EMAIL"),
		AuthPassword: os.Getenv("SMTP_AUTH_PASSWORD"),
	}

	return Mailer{
		emailConfig,
		"",
		nil,
	}
}

func (m Mailer) SendEmail(toEmail, subject string) Mailer {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("SMTP_SENDER_NAME")+" <"+m.emailConfig.AuthEmail+">")
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", m.Body)

	dialer := gomail.NewDialer(
		m.emailConfig.Host,
		m.emailConfig.Port,
		m.emailConfig.AuthEmail,
		m.emailConfig.AuthPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		m.Error = err
		return m
	}

	return m
}
