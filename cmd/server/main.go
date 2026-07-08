package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"dumb-proxy-go/internal/server-app"
	"dumb-proxy-go/pkg/wswrapper"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

func defaultRoute(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
			<html>
			<head><title>dumb-proxy-go</title></head>
			<body>type shi (v` + serverapp.Version + `)</body>
			</html>`))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("UPGRADE FAIL: %v", err)
		return
	}
	defer wsConn.Close()

	session, err := yamux.Server(wswrapper.NewGorillaConn(wsConn), nil)
	if err != nil {
		log.Printf("YAMUX SERVER FAIL: %v", err)
		return
	}
	defer session.Close()

	log.Println("MASTER TUNNEL OPEN!!! WAITING FOR MONKEY STREAMS... 🦍⚡")

	for {
		stream, err := session.Accept()
		if err != nil {
			log.Printf("SESSION CLOSED: %v", err)
			break
		}

		go handleVirtualStream(stream)
	}
}

func handleVirtualStream(stream net.Conn) {
	defer stream.Close()

	// Client has initiated new stream (e.g. it has opened youtube.com).
	//
	// Each stream is an isolated connection and every byte in that stream is necceserly
	// bound to target's destination. Client wants us to proxy that target, so we'll
	// read dst address is has passed us and allow flow for both ways.

	streamReader := io.Reader(stream)
	streamWriter := io.Writer(stream)

	line, err := bufio.NewReader(streamReader).ReadString('\n')
	if err != nil {
		log.Printf("READ FAIL %v", err)
		return
	}

	network, dst := line[:3], strings.TrimSpace(line[4:])

	if network == "tcp" {
		target, err := net.Dial("tcp", dst)
		if err != nil {
			log.Printf("DIAL TARGET FAIL %v", err)
			return
		}
		defer target.Close()

		targetReader := io.Reader(target)
		targetWriter := io.Writer(target)

		ch := make(chan error, 2)

		// Flow both ways
		go func() {
			// target <- stream
			_, err := io.Copy(targetWriter, streamReader)
			ch <- err
		}()

		go func() {
			// stream <- target
			_, err := io.Copy(streamWriter, targetReader)
			ch <- err
		}()

		<-ch
		return
	}

	target, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Printf("LISTEN PACKET FAIL %v", err)
		return
	}
	defer target.Close()

	ch := make(chan error, 2)

	// Flow both ways
	go func() {
		// target <- stream
		buf := make([]byte, 2048)

		for {
			n, err := streamReader.Read(buf)
			if err != nil {
				ch <- err
				return
			}

			if n == 0 {
				log.Printf("Error reading from target. empty frame")
				continue
			}

			b := buf[:n]

			idx := bytes.IndexByte(b, '\n')
			if idx == -1 {
				log.Printf("Error reading from stream. malformed frame")
				continue
			}

			line := string(b[:idx])
			addr, _ := net.ResolveUDPAddr(line[:3], strings.TrimSpace(line[4:]))

			log.Printf("WriteTo @ %s", addr.String())

			_, err = target.WriteTo(buf[len(line):n], addr)
			if err != nil {
				ch <- err
				return
			}
		}
	}()

	go func() {
		// stream <- target
		buf := make([]byte, 2048)

		for {
			n, addr, err := target.ReadFrom(buf)
			if err != nil {
				ch <- err
				return
			}

			if n == 0 {
				log.Printf("Error reading from target. empty packet")
				continue
			}

			b := []byte(fmt.Sprintf("udp %s\n", addr.String()))
			b = append(b, buf[:n]...)

			_, err = streamWriter.Write(b)
			if err != nil {
				ch <- err
				return
			}
		}
	}()

	<-ch
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	http.HandleFunc("/", defaultRoute)

	log.Println("SERVER RUNNING ON :8080... 🔥🔥🔥")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("PORT 8080 EXPLODED: %v", err)
	}
}
