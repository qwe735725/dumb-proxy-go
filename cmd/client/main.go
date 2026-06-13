package main

import (
	"context"
	"log"
	"net"

	"dumb-proxy-go/pkg/wswrapper"
	"github.com/armon/go-socks5"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

func main() {
	log.Println("MONKEY STARTING CLIENT... 🦍")

	wsUrl := "wss://dumb-proxy-go.onrender.com/ws"

	// 1. DIAL PORT 8080 FAST ⚡⚡
	wsConn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatalf("DIAL SMASHED!!! WEBSOCKET DEAD: %v", err)
	}
	defer wsConn.Close()
	log.Println("WEBSOCKET CONNECTED TO SERVER!!! 🔥🔥🔥")

	// 2. FORCE WEBSOCKET INTO NET.CONN 🛠️
	netConn := wswrapper.NewGorillaConn(wsConn)

	// 3. YAMUX MULTIPLEXER ON TOP 🌪️🌪️
	session, err := yamux.Client(netConn, nil)
	if err != nil {
		log.Fatalf("YAMUX CRASHED NO STREAM FOR YOU: %v", err)
	}
	defer session.Close()
	log.Println("YAMUX IS GO!!! MULTIPLEX MULTIPLEX MULTIPLEX!!! 🦍⚡")

	// 4. SOCKS5 HANDSHAKE ENGINE 🍌
	config := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			log.Printf("ME TOLD GO TO: %s 🍌", addr)

			// OPEN VIRTUAL STREAM
			stream, err := session.Open()
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
		log.Fatalf("SOCKS BUILDER BROKE: %v", err)
	}

	// 5. BIND TO 1080!!! 🦍🦖
	log.Println("SOCKS5 CLIENT RUNNING ON :1080!!! SEND DATA NOW!!! 🔥🔥🔥")
	if err := socksServer.ListenAndServe("tcp", ":1080"); err != nil {
		log.Fatalf("PORT 1080 EXPLODED: %v", err)
	}
}
