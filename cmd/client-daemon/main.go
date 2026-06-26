package main

import (
	"context"
	"flag"
	"log"
	"net"

	"dumb-proxy-go/internal/client-app"

	"github.com/armon/go-socks5"
	"github.com/pkg/errors"
)

func main() {
	log.Println("[🦍] MONKEY STARTING CLIENT...")

	wsUrl := flag.String("ws", "ws://localhost:8080/ws", "Remote proxy server WebSocket URL")
	flag.Parse()

	m := clientapp.NewMasterConn(*wsUrl)

	srv, err := socks5.New(&socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			log.Printf("[🍌] ME TOLD GO TO: %s", addr)

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
			_, err = stream.Write([]byte(addr + "\n"))
			if err != nil {
				stream.Close()
				return nil, err
			}

			return stream, nil
		},
	})
	if err != nil {
		log.Fatalf("[💥] SOCKS BUILDER BROKE: %v", err)
	}

	log.Println("[🦍] SOCKS5 CLIENT RUNNING ON :1080!!! SEND DATA NOW!!! 🔥🔥🔥")
	if err := srv.ListenAndServe("tcp", ":1080"); err != nil {
		log.Fatalf("[💥] PORT 1080 EXPLODED: %v", err)
	}
}
