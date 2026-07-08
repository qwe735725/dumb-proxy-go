package main

import (
	"bufio"
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
			<body>type shi (` + serverapp.Version + `)</body>
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

	target, err := net.Dial(network, dst)
	if err != nil {
		log.Printf("DIAL TARGET FAIL %v", err)
		return
	}
	defer target.Close()

	targetReader := io.Reader(target)
	targetWriter := io.Writer(target)

	// Flow both ways
	go func() {
		// target <- stream
		_, _ = io.Copy(targetWriter, streamReader)
	}()

	go func() {
		// stream <- target
		_, _ = io.Copy(streamWriter, targetReader)
	}()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	http.HandleFunc("/", defaultRoute)

	log.Println("SERVER RUNNING ON :8080... 🔥🔥🔥")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("PORT 8080 EXPLODED: %v", err)
	}
}
