package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"dumb-proxy-go/pkg/wswrapper"

	"github.com/armon/go-socks5"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
)

var (
	mx              sync.Mutex
	session         atomic.Pointer[yamux.Session]
	lastConnectedAt time.Time
)

func getSessionNonBlocking(wsURL string) (*yamux.Session, error) {
	s := session.Load()
	if s != nil && !s.IsClosed() {
		return s, nil
	}

	if !mx.TryLock() {
		return nil, errors.New("proxy is currently offline")
	}
	defer mx.Unlock()

	s = session.Load()
	if s != nil && !s.IsClosed() {
		return s, nil
	}

	if time.Since(lastConnectedAt) < 1*time.Second {
		return nil, errors.New("too many reconnects")
	}

	log.Println("[💥] WEBSOCKET SMASH!!! MASTER TUNNEL EXPLODE!!! RECONNECT WORKER ASSIGNED (🦍)!!!")

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	sess, err := yamux.Client(wswrapper.NewGorillaConn(wsConn), nil)
	if err != nil {
		wsConn.Close()
		return nil, err
	}

	session.Store(sess)
	lastConnectedAt = time.Now()

	log.Println("[🏆] RECONNECT WORKER (🦍) FIX TOTAL DESTRUCTION!!! MASTER TUNNEL RESTORED!!! GIVE BANANA!!! 🍌🍌🍌")
	return sess, nil
}

func main() {
	log.Println("[🦍] MONKEY STARTING CLIENT...")

	wsURL := flag.String("ws", "ws://localhost:8080/ws", "Remote proxy server WebSocket URL")
	flag.Parse()

	config := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			log.Printf("[🍌] ME TOLD GO TO: %s", addr)

			s, err := getSessionNonBlocking(*wsURL)
			if err != nil {
				return nil, err
			}

			// OPEN VIRTUAL STREAM
			stream, err := s.Open()
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
	}

	socksServer, err := socks5.New(config)
	if err != nil {
		log.Fatalf("[💥] SOCKS BUILDER BROKE: %v", err)
	}

	log.Println("[🦍] SOCKS5 CLIENT RUNNING ON :1080!!! SEND DATA NOW!!! 🔥🔥🔥")
	if err := socksServer.ListenAndServe("tcp", ":1080"); err != nil {
		log.Fatalf("[💥] PORT 1080 EXPLODED: %v", err)
	}
}
