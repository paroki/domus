package mailx_test

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedEmail struct {
	mailFrom string
	rcptTo   string
	auth     string
	data     string
	usedTLS  bool
}

type fakeSMTPServer struct {
	ln           net.Listener
	tlsCfg       *tls.Config
	offerSTARTLS bool
	delay        time.Duration
	mu           sync.Mutex
	received     []recordedEmail
}

func startFakeSMTPServer(t *testing.T, offerSTARTTLS bool, delay time.Duration) *fakeSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := &fakeSMTPServer{
		ln:           ln,
		tlsCfg:       selfSignedTLSConfig(t),
		offerSTARTLS: offerSTARTTLS,
		delay:        delay,
	}
	go srv.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return srv
}

func (s *fakeSMTPServer) port() int {
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *fakeSMTPServer) emails() []recordedEmail {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]recordedEmail, len(s.received))
	copy(out, s.received)
	return out
}

func (s *fakeSMTPServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *fakeSMTPServer) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeLine := func(line string) {
		_, _ = rw.WriteString(line + "\r\n")
		_ = rw.Flush()
	}

	writeLine("220 localhost ESMTP test")
	var rec recordedEmail
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO "), strings.HasPrefix(upper, "HELO "):
			writeLine("250-localhost")
			if s.offerSTARTLS && !rec.usedTLS {
				writeLine("250-STARTTLS")
			}
			writeLine("250 AUTH PLAIN")
		case upper == "STARTTLS":
			writeLine("220 Ready to start TLS")
			tlsConn := tls.Server(conn, s.tlsCfg)
			if err := tlsConn.Handshake(); err != nil {
				return
			}
			conn = tlsConn
			rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			rec.usedTLS = true
		case strings.HasPrefix(upper, "AUTH PLAIN "):
			raw := strings.TrimPrefix(line, line[:11])
			decoded, _ := base64.StdEncoding.DecodeString(raw)
			rec.auth = string(decoded)
			writeLine("235 2.7.0 Authentication successful")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			rec.mailFrom = strings.TrimSpace(line[len("MAIL FROM:"):])
			writeLine("250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			rec.rcptTo = strings.TrimSpace(line[len("RCPT TO:"):])
			writeLine("250 OK")
		case upper == "DATA":
			writeLine("354 End data with <CR><LF>.<CR><LF>")
			var body strings.Builder
			for {
				dline, derr := rw.ReadString('\n')
				if derr != nil {
					return
				}
				if strings.TrimRight(dline, "\r\n") == "." {
					break
				}
				body.WriteString(dline)
			}
			rec.data = body.String()
			writeLine("250 OK queued")
		case upper == "QUIT":
			writeLine("221 Bye")
			s.mu.Lock()
			s.received = append(s.received, rec)
			s.mu.Unlock()
			return
		default:
			writeLine("500 Unknown command")
		}
	}
}

func selfSignedTLSConfig(t *testing.T) *tls.Config {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	}
}

func TestMailerNew_Validation(t *testing.T) {
	_, err := mailx.New(mailx.Config{From: "Domus <no-reply@example.com>"})
	require.Error(t, err)

	_, err = mailx.New(mailx.Config{Host: "smtp.example.com", From: "invalid-from"})
	require.Error(t, err)

	m, err := mailx.New(mailx.Config{Host: "smtp.example.com", From: "Domus <no-reply@example.com>"})
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestMailerSend_PlainAuthAndHeaderSanitization(t *testing.T) {
	srv := startFakeSMTPServer(t, false, 0)
	m, err := mailx.New(mailx.Config{
		Host: "127.0.0.1",
		Port: srv.port(),
		User: "smtp-user",
		Pass: "smtp-pass",
		From: "Domus <no-reply@example.com>",
	})
	require.NoError(t, err)

	err = m.Send(context.Background(), mailx.Message{
		To:       "alice@example.com",
		Subject:  "Kode OTP\r\nBcc: evil@attacker.com",
		TextBody: "Kode Anda: 123456",
	})
	require.NoError(t, err)

	emails := srv.emails()
	require.Len(t, emails, 1)
	assert.Equal(t, "<no-reply@example.com>", emails[0].mailFrom)
	assert.Equal(t, "<alice@example.com>", emails[0].rcptTo)
	assert.Equal(t, "\x00smtp-user\x00smtp-pass", emails[0].auth)
	assert.Contains(t, emails[0].data, "Subject: Kode OTPBcc: evil@attacker.com\r\n")
	assert.NotContains(t, emails[0].data, "\r\nBcc:")
	assert.Contains(t, emails[0].data, "Kode Anda: 123456")
}

func TestMailerSend_ContextCancellation(t *testing.T) {
	srv := startFakeSMTPServer(t, false, 200*time.Millisecond)
	m, err := mailx.New(mailx.Config{
		Host: "127.0.0.1",
		Port: srv.port(),
		From: "Domus <no-reply@example.com>",
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err = m.Send(ctx, mailx.Message{
		To:       "alice@example.com",
		Subject:  "Timeout test",
		TextBody: "Hello",
	})
	require.Error(t, err)
	_ = strconv.Itoa(srv.port())
}
