package socks5

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

const (
	Socks5Version = 0x05

	AtypIPv4 = 0x01
	AtypFQDN = 0x03
	AtypIPv6 = 0x04
)

type Server struct {
	dial func(network, address string) (net.Conn, error)
}

func New(d func(network, address string) (net.Conn, error)) *Server {
	return &Server{
		dial: d,
	}
}

func (s *Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	// --- 1. Authentication Negotiation ---
	authReq := &ClientMethodSelectionMessage{}
	if err := authReq.FromReader(r); err != nil {
		return
	}

	authRes := ServerMethodSelectionMessage{ver: Socks5Version, method: AuthMethodNone}
	if _, err := conn.Write(authRes.Raw()); err != nil {
		return
	}

	// --- 2. Request Details ---
	req := &ClientRequestMessage{}
	if err := req.FromReader(r); err != nil {
		return
	}

	// --- 3. Execute Commands ---
	switch req.cmd {
	case CmdConnect:
		dst, err := addrString(req.atyp, req.dstAddr, req.dstPort)
		if err != nil {
			return
		}

		log.Printf("TCP Connect: %s", dst)

		s.handleConnect(conn, r, dst)

	case CmdBind:
		log.Printf("Bind")

		s.handleBind(conn)

	case CmdUdpAssociate:
		s.handleUdpAssociate(conn)

	default:
		// Command not supported
		m := &ServerReplyMessage{
			ver:     Socks5Version,
			rep:     RepNotAllowed,
			atyp:    AtypIPv4,
			bndAddr: make([]byte, 4),
		}
		conn.Write(m.Raw())
	}
}

// --- COMMAND HANDLERS ---

func (s *Server) handleConnect(conn net.Conn, r *bufio.Reader, dst string) {
	targetConn, err := s.dial("tcp", dst)
	if err != nil {
		m := &ServerReplyMessage{
			ver:     Socks5Version,
			rep:     RepServerFailure,
			atyp:    AtypIPv4,
			bndAddr: make([]byte, 4),
		}
		conn.Write(m.Raw())
		return
	}
	defer targetConn.Close()

	// Send success reply
	local := targetConn.LocalAddr().(*net.TCPAddr)

	m := &ServerReplyMessage{
		ver:     Socks5Version,
		rep:     RepSuccess,
		atyp:    AtypIPv4,
		bndAddr: local.IP,
		bndPort: uint16(local.Port),
	}
	if _, err := conn.Write(m.Raw()); err != nil {
		return
	}

	errChan := make(chan error, 2)

	go func() { _, err := io.Copy(targetConn, r); errChan <- err }()
	go func() { _, err := io.Copy(conn, targetConn); errChan <- err }()

	<-errChan
}

func (s *Server) handleBind(conn net.Conn) {
	// 1. Create a rendezvous TCP listener for the remote application to connect to
	bindListener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		m := &ServerReplyMessage{ver: Socks5Version, rep: RepServerFailure, atyp: AtypIPv4, bndAddr: make([]byte, 4)}
		conn.Write(m.Raw())
		return
	}
	defer bindListener.Close()

	bindIp := conn.LocalAddr().(*net.TCPAddr).IP
	bindPort := bindListener.Addr().(*net.TCPAddr).Port

	// Reply 1
	m := &ServerReplyMessage{
		ver:     Socks5Version,
		rep:     RepSuccess,
		atyp:    AtypIPv4,
		bndAddr: bindIp,
		bndPort: uint16(bindPort),
	}
	if _, err := conn.Write(m.Raw()); err != nil {
		return
	}

	// 2. Accept incoming connection from the target remote app
	targetConn, err := bindListener.Accept()
	if err != nil {
		m = &ServerReplyMessage{ver: Socks5Version, rep: RepServerFailure, atyp: AtypIPv4, bndAddr: make([]byte, 4)}
		conn.Write(m.Raw())
		return
	}
	defer targetConn.Close()

	// Reply 2
	if _, err := conn.Write(m.Raw()); err != nil {
		return
	}

	errChan := make(chan error, 2)

	go func() { _, err := io.Copy(targetConn, conn); errChan <- err }()
	go func() { _, err := io.Copy(conn, targetConn); errChan <- err }()

	<-errChan
}

func (s *Server) handleUdpAssociate(conn net.Conn) {
	// 1. Spin up a dynamic UDP listener for data relay
	udpListener, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		m := &ServerReplyMessage{
			ver:     Socks5Version,
			rep:     RepServerFailure,
			atyp:    AtypIPv4,
			bndAddr: make([]byte, 4),
		}
		conn.Write(m.Raw())
		return
	}
	defer udpListener.Close()

	bindIp := conn.LocalAddr().(*net.TCPAddr).IP.To4()
	bindPort := udpListener.LocalAddr().(*net.UDPAddr).Port

	log.Printf("[:%d] UDP", bindPort)

	// Reply: Inform client of our UDP port
	m := &ServerReplyMessage{
		ver:     Socks5Version,
		rep:     RepSuccess,
		atyp:    AtypIPv4,
		bndAddr: bindIp,
		bndPort: uint16(bindPort),
	}
	if _, err := conn.Write(m.Raw()); err != nil {
		return
	}

	remoteIp, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// 2. Start the UDP parsing/forwarding loop
	go func() {
		buf := make([]byte, 65535)
		var clientAddr *net.UDPAddr
		for {
			n, srcAddr, err := udpListener.ReadFrom(buf)
			if err != nil {
				return
			}

			srcIp, srcPortStr, _ := net.SplitHostPort(srcAddr.String())
			srcPort, _ := strconv.Atoi(srcPortStr)

			// If the packet comes from our expected client, forward it out
			if srcIp == remoteIp {
				clientAddr = srcAddr.(*net.UDPAddr)

				if n < 4 {
					continue
				}

				// Parse SOCKS5 UDP Header
				// RSV (2B), FRAG (1B), ATYP (1B)

				h := &UdpPacket{}
				if err := h.FromReader(bytes.NewReader(buf[:n])); err != nil {
					continue
				}

				targetAddrStr, err := addrString(h.atyp, h.dstAddr, h.dstPort)
				if err != nil {
					continue
				}

				log.Printf("[:%d] FWD to %s", bindPort, targetAddrStr)

				// Send raw payload to final destination
				if remoteUDP, err := s.dial("udp", targetAddrStr); err == nil {
					remoteUDP.Write(h.data)
					remoteUDP.Close()
				}
				continue
			}

			if clientAddr == nil {
				continue
			}

			// Build SOCKS5 UDP encapsulation header
			h := &UdpPacket{
				atyp:    AtypIPv4,
				dstAddr: net.ParseIP(srcIp).To4(),
				dstPort: uint16(srcPort),
				data:    buf[:n],
			}

			if h.dstAddr == nil {
				h.atyp = AtypIPv6
				h.dstAddr = net.ParseIP(srcIp).To16()
			}

			udpListener.WriteTo(h.Raw(), clientAddr)
			log.Printf("[:%d] BACK to %s", bindPort, clientAddr.String())
		}
	}()

	// Keep UDP open as long as the primary TCP command session remains alive
	dummy := make([]byte, 512)
	for {
		_, err := conn.Read(dummy)
		if err != nil {
			break // Client terminated the TCP control connection; tear down UDP listener
		}
	}

	log.Printf("[:%d] connection closed.", bindPort)
}

// --- HELPERS ---
func addrString(atyp uint8, addr []byte, port uint16) (string, error) {
	host := ""
	switch atyp {
	case AtypIPv4, AtypIPv6:
		host = net.IP(addr).String()

	case AtypFQDN:
		if len(addr) < 2 {
			return "", errors.New("bad FQDN address")
		}
		host = string(addr[1:])

	default:
		return "", errors.New("unknown address type")
	}

	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func readAddr(r io.Reader, atyp byte) ([]byte, error) {
	switch atyp {
	case AtypIPv4:
		b := make([]byte, 4)
		_, err := io.ReadFull(r, b)
		return b, err

	case AtypFQDN:
		b := make([]byte, 1)
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}

		b = append(b, make([]byte, b[0])...)

		_, err := io.ReadFull(r, b[1:])
		return b, err

	case AtypIPv6:
		b := make([]byte, 16)
		_, err := io.ReadFull(r, b)
		return b, err
	}

	return nil, errors.New("unsupported address structure mapping")
}
