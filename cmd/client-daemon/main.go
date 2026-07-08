package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"time"

	"dumb-proxy-go/internal/client-app"

	"dumb-proxy-go/pkg/socks5"
	"github.com/pkg/errors"
)

type readResult struct {
	n    int
	addr net.Addr
	err  error

	p []byte
}

type GorillaUdpConn struct {
	clientAddr string

	local  net.PacketConn
	remote net.Conn

	localPending  chan readResult
	remotePending chan readResult

	isClosed bool
}

func NewGorillaUdpConn(remote net.Conn, clientAddr string, network, address string) (net.PacketConn, error) {
	local, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, err
	}

	c := &GorillaUdpConn{
		clientAddr: clientAddr,

		local:  local,
		remote: remote,

		remotePending: make(chan readResult),
		localPending:  make(chan readResult),
	}

	go func() {
		buf := make([]byte, 2048)

		for !c.isClosed {
			n, err := c.remote.Read(buf)
			if n == 0 {
				log.Printf("Error reading from remote. empty frame")
				continue
			}

			b := buf[:n]
			idx := bytes.IndexByte(b, '\n')
			if idx == -1 {
				log.Printf("Error reading from remote. malformed frame")
				continue
			}


			line := string(b[:idx])
			addr, _ := net.ResolveUDPAddr(line[:3], strings.TrimSpace(line[4:]))

			log.Printf("remote read: %s", addr.String())

			c.remotePending <- readResult{n: n - len(line), addr: addr, err: err, p: slices.Clone(buf[len(line):n])}
		}
	}()

	go func() {
		buf := make([]byte, 2048)

		for !c.isClosed {
			n, addr, err := c.local.ReadFrom(buf)
			if n == 0 {
				log.Printf("Error reading from local. empty packet")
				continue
			}

			c.localPending <- readResult{n: n, addr: addr, err: err, p: slices.Clone(buf[:n])}
		}
	}()

	return c, nil
}

func (c *GorillaUdpConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	select {
	case r := <-c.localPending:
		copy(p, r.p)
		return r.n, r.addr, r.err
	case r := <-c.remotePending:
		copy(p, r.p)
		return r.n, r.addr, r.err
	}
}

func (c *GorillaUdpConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if addr.String() == c.clientAddr {
		return c.local.WriteTo(p, addr)
	}

	b := []byte(fmt.Sprintf("udp %s\n", addr.String()))
	b = append(b, p...)

	return c.remote.Write(b)
}

func (c *GorillaUdpConn) Close() error {
	c.isClosed = true

	err1 := c.remote.Close()
	err2 := c.local.Close()

	if err1 == nil && err2 == nil {
		return nil
	}

	return errors.Errorf("%v %v", err1, err2)
}

func (c *GorillaUdpConn) LocalAddr() net.Addr {
	return c.local.LocalAddr()
}

func (c *GorillaUdpConn) SetDeadline(t time.Time) error { return nil }

func (c *GorillaUdpConn) SetReadDeadline(t time.Time) error { return nil }

func (c *GorillaUdpConn) SetWriteDeadline(t time.Time) error { return nil }

func main() {
	log.Println("[🦍] MONKEY STARTING CLIENT...")

	wsUrl := flag.String("ws", "ws://localhost:8080/ws", "Remote proxy server WebSocket URL")
	flag.Parse()

	m := clientapp.NewMasterConn(*wsUrl)

	d := func(network, addr string) (net.Conn, error) {
		conn := m.YamuxConn()
		if conn == nil || conn.IsClosed() {
			m.TriggerReconnect()
			return nil, errors.New("proxy is currently offline")
		}

		// OPEN VIRTUAL STREAM
		stream, err := conn.Open()
		if err != nil {
			return nil, err
		}

		// TELL SERVER WHERE TO GO
		_, err = stream.Write([]byte(network + " " + addr + "\n"))
		if err != nil {
			stream.Close()
			return nil, err
		}

		return stream, nil
	}

	l := func(ctx context.Context, network, address string) (net.PacketConn, error) {
		conn := m.YamuxConn()
		if conn == nil || conn.IsClosed() {
			m.TriggerReconnect()
			return nil, errors.New("proxy is currently offline")
		}

		// OPEN VIRTUAL STREAM
		stream, err := conn.Open()
		if err != nil {
			return nil, err
		}

		// TELL SERVER WHERE TO GO
		_, err = stream.Write([]byte(network + " " + address + "\n"))
		if err != nil {
			stream.Close()
			return nil, err
		}

		return NewGorillaUdpConn(stream, ctx.Value("clientAddress").(string), network, address)
	}

	srv := socks5.New(d, l)

	log.Println("[🦍] SOCKS5 CLIENT RUNNING ON :1080!!! SEND DATA NOW!!! 🔥🔥🔥")
	if err := srv.ListenAndServe("tcp", ":1080"); err != nil {
		log.Fatalf("[💥] PORT 1080 EXPLODED: %v", err)
	}
}
