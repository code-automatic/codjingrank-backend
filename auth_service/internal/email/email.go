// Package email delivers transactional authentication emails over SMTP.
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"innogen-backend/shared/config"
)

// Sender sends password-reset one-time codes over SMTP with STARTTLS.
type Sender struct {
	host string
	port string
	user string
	pass string
	from string
	log  *slog.Logger
}

// New creates a Sender from configuration. An empty SMTP_HOST disables real
// delivery, which keeps development and test environments unblocked.
func New(cfg *config.Config, log *slog.Logger) *Sender {
	return &Sender{
		host: cfg.SMTPHost,
		port: cfg.SMTPPort,
		user: cfg.SMTPUser,
		pass: cfg.SMTPPass,
		from: cfg.SMTPFrom,
		log:  log,
	}
}

// SendOTP emails a six-digit reset code. When SMTP is not configured the call
// logs a warning and returns nil so callers can keep a uniform response.
func (s *Sender) SendOTP(ctx context.Context, to, otp string, ttl time.Duration) error {
	if strings.TrimSpace(s.host) == "" {
		s.log.Warn("SMTP is not configured; skipping password reset email",
			slog.String("recipient", to))
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("email.SendOTP: invalid recipient")
	}

	from := strings.TrimSpace(s.from)
	if from == "" {
		from = strings.TrimSpace(s.user)
	}
	if from == "" {
		return fmt.Errorf("email.SendOTP: SMTP_FROM is not configured")
	}
	if strings.ContainsAny(from, "\r\n") {
		return fmt.Errorf("email.SendOTP: invalid sender")
	}

	port := strings.TrimSpace(s.port)
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(s.host, port)

	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("email.SendOTP dial: %w", err)
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return fmt.Errorf("email.SendOTP deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("email.SendOTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("email.SendOTP: SMTP server does not support STARTTLS")
	}
	if err = client.StartTLS(&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}); err != nil {
		return fmt.Errorf("email.SendOTP starttls: %w", err)
	}
	if s.user != "" {
		if err = client.Auth(smtp.PlainAuth("", s.user, s.pass, s.host)); err != nil {
			return fmt.Errorf("email.SendOTP auth: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("email.SendOTP mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("email.SendOTP rcpt to: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("email.SendOTP data: %w", err)
	}
	if _, err = writer.Write(buildMessage(from, to, otp, ttl)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("email.SendOTP write: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("email.SendOTP close body: %w", err)
	}
	if err = client.Quit(); err != nil {
		return fmt.Errorf("email.SendOTP quit: %w", err)
	}
	return nil
}

func buildMessage(from, to, otp string, ttl time.Duration) []byte {
	minutes := int(ttl.Round(time.Minute) / time.Minute)
	if minutes <= 0 {
		minutes = 30
	}
	subject := mime.QEncoding.Encode("UTF-8", "Mã OTP đặt lại mật khẩu Codjingrank")
	body := fmt.Sprintf("Mã OTP đặt lại mật khẩu Codjingrank của bạn là: %s. Mã có hiệu lực %d phút.", otp, minutes)

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return []byte(b.String())
}
