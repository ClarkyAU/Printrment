package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Optional HTTPS for server mode (Docker / TrueNAS). Without it the admin
// password and the phone PIN cross your network as plain text, which is only
// acceptable on a network you trust. Turn it on with PRINTRMENT_HTTPS=1: the
// app makes itself a certificate (browsers will warn once, because nobody
// else vouches for it), or use your own with PRINTRMENT_TLS_CERT and
// PRINTRMENT_TLS_KEY (for example one from Tailscale or your own CA).

func httpsEnabled() bool {
	return serverMode && (os.Getenv("PRINTRMENT_HTTPS") == "1" ||
		(os.Getenv("PRINTRMENT_TLS_CERT") != "" && os.Getenv("PRINTRMENT_TLS_KEY") != ""))
}

func scheme() string {
	if httpsEnabled() {
		return "https"
	}
	return "http"
}

func tlsDir() string {
	d := filepath.Join(os.Getenv("PRINTRMENT_DATA"), "key")
	os.MkdirAll(d, 0o700)
	return d
}

// serverTLS returns the TLS settings for the server's listeners.
func serverTLS() (*tls.Config, error) {
	certFile, keyFile := os.Getenv("PRINTRMENT_TLS_CERT"), os.Getenv("PRINTRMENT_TLS_KEY")
	if certFile == "" || keyFile == "" {
		certFile, keyFile = filepath.Join(tlsDir(), "https-cert.pem"), filepath.Join(tlsDir(), "https-key.pem")
		if !usableCert(certFile, keyFile) {
			if err := makeSelfSigned(certFile, keyFile); err != nil {
				return nil, err
			}
		}
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("can't load the HTTPS certificate: %v", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

// usableCert is true for an existing self-made certificate with more than 30
// days left.
func usableCert(certFile, keyFile string) bool {
	if _, err := os.Stat(keyFile); err != nil {
		return false
	}
	b, err := os.ReadFile(certFile)
	if err != nil {
		return false
	}
	blk, _ := pem.Decode(b)
	if blk == nil {
		return false
	}
	c, err := x509.ParseCertificate(blk.Bytes)
	return err == nil && time.Until(c.NotAfter) > 30*24*time.Hour
}

func makeSelfSigned(certFile, keyFile string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	host, _ := os.Hostname()
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{"Printrment"}, CommonName: "Printrment"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	if host != "" {
		tpl.DNSNames = append(tpl.DNSNames, host)
	}
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() {
			tpl.IPAddresses = append(tpl.IPAddresses, n.IP)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}), 0o600); err != nil {
		return err
	}
	return os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644)
}

// serve runs srv on ln, over HTTPS when server mode has it turned on.
func serve(srv *http.Server, ln net.Listener) {
	if httpsEnabled() {
		if conf, err := serverTLS(); err != nil {
			appendLog("HTTPS: " + err.Error())
			fmt.Fprintln(os.Stderr, "Printrment: HTTPS:", err)
			os.Exit(1)
		} else {
			ln = tls.NewListener(ln, conf)
		}
	}
	srv.Serve(ln)
}

// securityHeaders are added to every response. The page's own script and
// style are inline, so those two allow it; everything else is locked to this
// origin, and the page can't be framed, can't load plugins or change <base>.
func securityHeaders(h http.Header) {
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; "+
		"img-src 'self' data: blob:; media-src 'self' blob:; connect-src 'self'; font-src 'self'; "+
		"object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
}
