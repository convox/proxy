package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func die(err error) {
	warn(err)
	os.Exit(1)
}

func warn(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: proxy <from> <to> <protocol> [options]\n")
		os.Exit(1)
	}

	from := os.Args[1]
	to := os.Args[2]
	protocol := os.Args[3]
	proxy := false
	secure := false

	for _, option := range os.Args[4:] {
		switch option {
		case "proxy":
			proxy = true
		case "secure":
			secure = true
		}
	}

	if len(os.Args) > 4 && os.Args[4] == "proxy" {
		proxy = true
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", from))

	if err != nil {
		die(err)
	}

	switch protocol {
	case "https", "tls":
		cert, err := generateSelfSignedCertificate("convox.local")

		if err != nil {
			die(err)
		}

		ln = tls.NewListener(ln, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
	}

	defer ln.Close()

	fmt.Printf("listen %s\n", from)

	for {
		conn, err := ln.Accept()

		if err != nil {
			die(err)
		}

		if proxy {
			go handleProxyConnection(conn, to, secure)
		} else {
			go handleTcpConnection(conn, to, secure)
		}
	}
}

func dial(addr string, secure bool) (net.Conn, error) {
	if secure {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}

		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		return tls.DialWithDialer(dialer, "tcp", addr, config)
	} else {
		return net.DialTimeout("tcp", addr, 5*time.Second)
	}

}

func handleProxyConnection(in net.Conn, to string, secure bool) {
	rp := strings.SplitN(in.RemoteAddr().String(), ":", 2)
	top := strings.SplitN(to, ":", 2)

	fmt.Printf("proxy %s:%s -> %s:%s secure=%t\n", rp[0], rp[1], top[0], top[1], secure)

	out, err := dial(to, secure)

	if err != nil {
		warn(err)
		return
	}

	header := fmt.Sprintf("PROXY TCP4 %s 127.0.0.1 %s %s\r\n", rp[0], rp[1], top[1])

	out.Write([]byte(header))

	pipe(in, out)
}

func handleTcpConnection(in net.Conn, to string, secure bool) {
	rp := strings.SplitN(in.RemoteAddr().String(), ":", 2)
	top := strings.SplitN(to, ":", 2)

	fmt.Printf("tcp %s:%s -> %s:%s secure=%t\n", rp[0], rp[1], top[0], top[1], secure)

	out, err := dial(to, secure)

	if err != nil {
		warn(err)
		return
	}

	pipe(in, out)
}

func pipe(a, b io.ReadWriter) {
	var wg sync.WaitGroup

	wg.Add(2)
	go copyWait(a, b, &wg)
	go copyWait(b, a, &wg)
	wg.Wait()
}

func copyWait(to io.Writer, from io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(to, from)
}

func generateSelfSignedCertificate(host string) (tls.Certificate, error) {
	rkey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"convox"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	data, err := x509.CreateCertificate(rand.Reader, &template, &template, &rkey.PublicKey, rkey)

	if err != nil {
		return tls.Certificate{}, err
	}

	pub := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rkey)})

	return tls.X509KeyPair(pub, key)
}
