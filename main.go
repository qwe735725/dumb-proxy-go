package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net/http"
	"time"
	"io"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

func main() {
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			fmt.Println("Wrong content type type shi")
			ok(w)
			return
		}

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error reading body type shi: %s", err.Error())
			return
		}
		defer r.Body.Close()

		/*packet :=*/
		gopacket.NewPacket(bytes, layers.LayerTypeIPv4, gopacket.Default)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
				<html>
				<head><title>dumb-proxy-go</title></head>
				<body>/proxy</body>
				</html>`))
		//route(packet)
		// ip.dst : tcp.dstPort
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 2. Automatically handle the handshake protocol upgrade
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		log.Println("WebSocket link established via Gorilla!")

		// 3. Server-to-Client Stream Loop (Outbound)
		go func() {
			count := 0
			for {
				count++
				payload := fmt.Sprintf("Gorilla Server Tick #%d | Time: %s", count, time.Now().Format("15:04:05"))
				
				// Simple native helper to send string messages safely
				err := conn.WriteMessage(websocket.TextMessage, []byte(payload))
				if err != nil {
					return // Triggers if client disconnects
				}
				time.Sleep(2 * time.Second)
			}
		}()

		// 4. Client-to-Server Stream Loop (Inbound)
		for {
			// Automatically handles unmasking and structural frame management
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Client dropped connection.")
				break
			}

			log.Printf("Received from Client: %s\n", string(message))

			// Echo receipt confirmation straight back over the socket
			ack := fmt.Sprintf("Server Acknowledged: '%s'", string(message))
			if err := conn.WriteMessage(messageType, []byte(ack)); err != nil {
				break
			}
		}
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>WebSocket Stream Test</title>
			<style>
				body { font-family: sans-serif; margin: 40px; background: #f4f4f9; color: #333; }
				#log { background: #1e1e1e; color: #76c176; padding: 20px; border-radius: 6px; font-family: monospace; white-space: pre-wrap; height: 350px; overflow-y: auto; margin-bottom: 20px;}
				h1 { color: #2c3e50; }
				button { padding: 10px 20px; background: #3498db; color: #fff; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; }
				button:hover { background: #2980b9; }
			</style>
		</head>
		<body>
			<h1>Bi-Directional WebSocket Stream Inspector</h1>
			<p>Watch incoming server ticks below, or click the button to stream client chunks back up:</p>
			
			<div id="log">Initializing WebSocket upgrade request...&#10;</div>
			<button id="sendBtn" disabled>Stream Chunk to Server</button>

			<script>
				let ws;
				let clientCount = 0;

				function initWebSocket() {
					const logDiv = document.getElementById('log');
					const sendBtn = document.getElementById('sendBtn');

					// 1. Open connection using the ws:// protocol (bypasses load balancer HTTP logic)
					ws = new WebSocket('ws://' + window.location.host + '/ws');

					ws.onopen = () => {
						logDiv.textContent += "WebSocket connection opened! Buffering disabled.\n\n";
						sendBtn.disabled = false; // Allow client communication
					};

					// 2. Event listener for server data streams
					ws.onmessage = (event) => {
						logDiv.textContent += "[Inbound] " + event.data + "\n";
						logDiv.scrollTop = logDiv.scrollHeight;
					};

					ws.onclose = () => {
						logDiv.textContent += "\n[WebSocket link severed]";
						sendBtn.disabled = true;
					};
				}

				// 3. User interaction pushes a real-time event block up to the server stream
				document.getElementById('sendBtn').onclick = () => {
					clientCount++;
					const payload = "Client Frame Element #" + clientCount;
					
					ws.send(payload); // Dispatched instantly through the open pipe
					
					const logDiv = document.getElementById('log');
					logDiv.textContent += "[Outbound] Sent: " + payload + "\n";
					logDiv.scrollTop = logDiv.scrollHeight;
				};

				initWebSocket();
			</script>
		</body>
		</html>
		`
		w.Write([]byte(html))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ok(w)
	})

	fmt.Println("Server running on :8080...")
	http.ListenAndServe(":8080", nil)
}
