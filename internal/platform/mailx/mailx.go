package mailx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Config holds SMTP connection and sender settings.
type Config struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

// Message represents a plain-text email message.
type Message struct {
	To       string
	Subject  string
	TextBody string
}

// Mailer sends emails over SMTP using Go standard library primitives.
type Mailer struct {
	cfg        Config
	fromAddr   string
	fromHeader string
}

// New validates cfg and constructs a Mailer.
func New(cfg Config) (*Mailer, error) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	if cfg.Host == "" {
		return nil, errors.New("mailx: smtp host is required")
	}
	if cfg.Port <= 0 {
		cfg.Port = 587
	}
	rawFrom := strings.TrimSpace(cfg.From)
	addr, err := mail.ParseAddress(rawFrom)
	if err != nil || addr.Address == "" {
		return nil, fmt.Errorf("mailx: invalid from address %q: %w", cfg.From, err)
	}
	return &Mailer{
		cfg:        cfg,
		fromAddr:   addr.Address,
		fromHeader: sanitizeHeader(rawFrom),
	}, nil
}

// Send delivers msg over SMTP respecting ctx cancellation and deadlines.
func (m *Mailer) Send(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	rawTo := strings.TrimSpace(msg.To)
	toAddr, err := mail.ParseAddress(rawTo)
	if err != nil || toAddr.Address == "" {
		return fmt.Errorf("mailx: invalid recipient address %q: %w", msg.To, err)
	}

	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("mailx: dial %s: %w", addr, err)
	}

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(15 * time.Second)
	}
	_ = conn.SetDeadline(deadline)

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	var client *smtp.Client
	if m.cfg.Port == 465 {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: m.cfg.Host,
			MinVersion: tls.VersionTLS12,
		})
		client, err = smtp.NewClient(tlsConn, m.cfg.Host)
	} else {
		client, err = smtp.NewClient(conn, m.cfg.Host)
	}
	if err != nil {
		_ = conn.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("mailx: smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if m.cfg.Port != 465 && !isLoopbackHost(m.cfg.Host) {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{
				ServerName: m.cfg.Host,
				MinVersion: tls.VersionTLS12,
			}); err != nil {
				return fmt.Errorf("mailx: starttls: %w", err)
			}
		}
	}

	if m.cfg.User != "" || m.cfg.Pass != "" {
		auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("mailx: auth: %w", err)
		}
	}

	if err := client.Mail(m.fromAddr); err != nil {
		return fmt.Errorf("mailx: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toAddr.Address); err != nil {
		return fmt.Errorf("mailx: RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailx: DATA: %w", err)
	}

	payload := buildMessage(m.fromHeader, sanitizeHeader(rawTo), sanitizeHeader(msg.Subject), msg.TextBody)
	if _, err := w.Write([]byte(payload)); err != nil {
		_ = w.Close()
		return fmt.Errorf("mailx: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailx: close data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("mailx: quit: %w", err)
	}
	return nil
}

func sanitizeHeader(v string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(strings.TrimSpace(v))
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func buildMessage(from, to, subject, textBody string) string {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\nTo: ")
	b.WriteString(to)
	b.WriteString("\r\nSubject: ")
	b.WriteString(subject)
	b.WriteString("\r\nDate: ")
	b.WriteString(time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(textBody)
	return b.String()
}
