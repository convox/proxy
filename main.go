package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func die(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: proxy <from> <to> <protocol>\n")
		os.Exit(1)
	}

	from := os.Args[1]
	to := os.Args[2]
	protocol := os.Args[3]

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", from))

	if err != nil {
		die(err)
	}

	defer ln.Close()

	fmt.Printf("listen %s\n", from)

	for {
		conn, err := ln.Accept()

		if err != nil {
			die(err)
		}

		switch protocol {
		case "http":
			go handleHttpConnection(conn, to)
		case "proxy":
			go handleProxyConnection(conn, to)
		case "tcp":
			go handleTcpConnection(conn, to)
		default:
			die(fmt.Errorf("unknown protocol: %s", protocol))
		}
	}
}

func handleHttpConnection(in net.Conn, to string) {
	handleTcpConnection(in, to)
}

func handleProxyConnection(in net.Conn, to string) {
	rp := strings.SplitN(in.RemoteAddr().String(), ":", 2)
	top := strings.SplitN(to, ":", 2)

	fmt.Printf("proxy %s:%s -> %s:%s\n", rp[0], rp[1], top[0], top[1])

	out, err := net.DialTimeout("tcp", to, 5*time.Second)

	if err != nil {
		die(err)
	}

	header := fmt.Sprintf("PROXY TCP4 %s 127.0.0.1 %s %s\r\n", rp[0], rp[1], top[1])

	out.Write([]byte(header))

	pipe(in, out)
}

func handleTcpConnection(in net.Conn, to string) {
	rp := strings.SplitN(in.RemoteAddr().String(), ":", 2)
	top := strings.SplitN(to, ":", 2)

	fmt.Printf("tcp %s:%s -> %s:%s\n", rp[0], rp[1], top[0], top[1])

	out, err := net.DialTimeout("tcp", to, 5*time.Second)

	if err != nil {
		die(err)
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
