package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net/http"
	"time"
	"io"
)

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

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming or HTTP/2 not supported by network connection", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		log.Println("New client connected to /stream")

		count := 0
		for {
			select {
			case <-ctx.Done():
				// This fires instantly if the user closes the window or navigates away
				log.Printf("Client disconnected from stream. Total chunks sent: %d\n", count)
				return
			default:
				count++
				// Format and write the data payload
				payload := fmt.Sprintf("HTTP/2 Data Frame Chunk #%d | Timestamp: %s\n", count, time.Now().Format("15:04:05"))
				_, err := w.Write([]byte(payload))
				if err != nil {
					log.Printf("Write error: %v\n", err)
					return
				}

				// Force the HTTP/2 engine to transmit the DATA frame immediately over the wire
				flusher.Flush()

				// Sleep to simulate pacing real-time background events
				time.Sleep(1 * time.Second)
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
			<title>HTTP/2 Stream Test</title>
			<style>
				body { font-family: sans-serif; margin: 40px; background: #f4f4f9; color: #333; }
				#log { background: #1e1e1e; color: #76c176; padding: 20px; border-radius: 6px; font-family: monospace; white-space: pre-wrap; height: 400px; overflow-y: auto; }
				h1 { color: #2c3e50; }
			</style>
		</head>
		<body>
			<h1>HTTP/2 Infinite Stream Inspector</h1>
			<p>Open your browser DevTools Console (F12) or watch the terminal block below:</p>
			<div id="log">Connecting to stream...&#10;</div>

			<script>
				async function listenToHttp2Stream() {
					const logDiv = document.getElementById('log');
					
					try {
						// Fetch handles HTTP/2 connection pooling automatically
						const response = await fetch('/stream');
						const reader = response.body.getReader();
						const decoder = new TextDecoder();

						logDiv.textContent += "Connected successfully! Reading binary frames... \n\n";

						while (true) {
							// reader.read() resolves exactly when an HTTP/2 DATA frame lands
							const { value, done } = await reader.read();
							if (done) {
								logDiv.textContent += "\n[Stream ended by server]";
								break;
							}

							const chunkText = decoder.decode(value);
							
							// Print to standard browser devtools console
							console.log("Intercepted Chunk:", chunkText.trim());
							
							// Append to the UI display container
							logDiv.textContent += chunkText;
							logDiv.scrollTop = logDiv.scrollHeight; // Auto-scroll to bottom
						}
					} catch (error) {
						logDiv.textContent += "\n[Network Error]: " + error.message;
						console.error("Stream reader broken:", error);
					}
				}

				// Fire stream ingestion immediately on page load
				listenToHttp2Stream();
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
