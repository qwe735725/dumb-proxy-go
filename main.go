package main

import (
	"bufio"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

func ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
			<html>
			<head><title>dumb-proxy-go</title></head>
			<body>type shi</body>
			</html>`))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. UPGRADE HTTP CONNECTION TO GORILLA WEBSOCKET 🦍
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("UPGRADE FAIL: %v", err)
		return
	}
	defer wsConn.Close()

	// 2. EXTRACT RAW NETWORK CONNECTION VIA UNDERLYING NATIVE SOCKET 🛠️
	netConn := NewGorillaConn(wsConn)

	// 3. SLAP YAMUX SERVER ON TOP TO UNWRAP STREAMS 🌪️🌪️
	session, err := yamux.Server(netConn, nil)
	if err != nil {
		log.Printf("YAMUX SERVER FAIL: %v", err)
		return
	}
	defer session.Close()

	log.Println("MASTER TUNNEL OPEN!!! WAITING FOR MONKEY STREAMS... 🦍⚡")

	// 4. LOOP FOREVER ACCEPTING VIRTUAL STREAMS INSIDE WEBSOCKET
	for {
		stream, err := session.Accept()
		if err != nil {
			log.Printf("SESSION CLOSED: %v", err)
			break
		}

		// PROCESS EACH STREAM CONCURRENTLY FAST!!! 🔥
		go handleVirtualStream(stream)
	}
}

func handleVirtualStream(stream net.Conn) {
	defer stream.Close()

	// 1. READ TARGET ADDRESS STRING FROM CLIENT UNTIL NEWLINE 🍌
	reader := bufio.NewReader(stream)
	targetAddr, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("BAD READ: %v", err)
		return
	}
	targetAddr = strings.TrimSpace(targetAddr)

	log.Printf("DIALING INTERNET TARGET: %s 🚀", targetAddr)

	// 2. DIAL OUT TO REAL WEBSITE ON THE INTERNET
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("DIAL TARGET FAIL %s: %v", targetAddr, err)
		return
	}
	defer targetConn.Close()

	// 3. BI-DIRECTIONAL RAW BYTES COPY PIPE!!! ZERO DECODING!!! 🌪️🚀
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(targetConn, reader) // CLIENT TO INTERNET
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(stream, targetConn) // INTERNET TO CLIENT
		errChan <- err
	}()

	// WAIT UNTIL ONE SIDE HANGS UP OR RESETS
	<-errChan
	log.Printf("TARGET CONNECTION %s CLOSED CLEANLY!!! 🎉", targetAddr)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ok(w)
	})

	log.Println("SERVER RUNNING ON :8080... 🔥🔥🔥")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("PORT 8080 EXPLODED: %v", err)
	}
}
