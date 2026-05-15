package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

type Service struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewService() *Service {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "Undangan Digital <noreply@undangan-digital.anggriawan.my.id>"
	}
	return &Service{
		host:     os.Getenv("SMTP_HOST"),
		port:     port,
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASS"),
		from:     from,
	}
}

func (s *Service) SendOTP(toEmail, name, otp string) error {
	subject := "Kode Verifikasi Undangan Digital"
	body := fmt.Sprintf(`Halo %s!

Kode verifikasi kamu:

  ======
  %s
  ======

Kode berlaku 10 menit. Jangan bagikan ke siapapun.

Salam,
Tim Undangan Digital`, name, otp)

	return s.send(toEmail, subject, body)
}

func (s *Service) SendPasswordReset(toEmail, name, resetLink string) error {
	subject := "Reset Password Undangan Digital"
	body := fmt.Sprintf(`Halo %s,

Kamu meminta reset password. Klik link di bawah:

%s

Link berlaku 1 jam. Jika tidak merasa meminta, abaikan email ini.

Salam,
Tim Undangan Digital`, name, resetLink)

	return s.send(toEmail, subject, body)
}

func (s *Service) send(to, subject, body string) error {
	if s.host == "" {
		// SMTP not configured, log only (dev mode)
		fmt.Printf("[EMAIL] To: %s\nSubject: %s\n%s\n---\n", to, subject, body)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body)

	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	fromAddr := s.user
	if fromAddr == "" {
		fromAddr = s.from
	}
	return smtp.SendMail(addr, auth, fromAddr, []string{to}, []byte(msg))
}
