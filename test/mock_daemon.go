package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	udsPath := "/tmp/dumb-proxy.sock"
	_ = os.Remove(udsPath) // Clear any leftover dead file locks from previous crashes

	listener, err := net.Listen("unix", udsPath)
	if err != nil {
		log.Fatalf("[-] Failed to boot mock daemon engine: %v", err)
	}
	defer os.Remove(udsPath)

	// 🛡️ THE IMMORTAL FORCE SHUTDOWN MONITOR
	// Capture standard Ctrl+C interrupts so we can wipe the socket file cleanly before dying
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n[🧹] Wiping Unix file socket. Mock daemon shut down cleanly.")
		_ = os.Remove(udsPath)
		os.Exit(0)
	}()

	log.Println("[🦍] TELEMETRY SIMULATOR RUNNING PERMANENTLY IN THE TREES... 🔥")
	log.Println("[ℹ️] Press Ctrl+C in this terminal window to kill the process.")

	// =========================================================================
	// 🔄 THE INFINITE CONNECTION ACCEPTER BLOCK
	// =========================================================================
	for {
		log.Println("[⏳] Waiting for a dashboard client to attach to the socket...")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[-] Listener error: %v. Retrying...", err)
			continue
		}
		
		log.Println("[🎉] Dashboard attached! Launching sub-thread to fire live data stream...")
		
		// Spawn a dedicated thread handler for this specific dashboard session
		// This leaves the main loop completely free to instantly pick up the next socket dial!
		handleDashboardSession(conn)
	}
}

func handleDashboardSession(conn net.Conn) {
	defer conn.Close()

	targets := []string{"github.com:443", "discord.com:443", "youtube.com:443", "reddit.com:443"}
	var active []string
	var rxVol, txVol float64 = 142.5, 12.4
	tick := 0

	for {
		time.Sleep(500 * time.Millisecond)
		tick++

		// Simulate dynamic stream counts shifting up and down over time
		if tick%6 == 1 && len(active) < len(targets) {
			active = append(active, targets[len(active)])
		} else if tick%10 == 0 && len(active) > 0 {
			active = active[:len(active)-1]
		}

		rxSpeed, txSpeed := 0, 0
		if len(active) > 0 {
			rxSpeed = 300 + (tick%4)*45
			txSpeed = 12 + (tick%3)*3
			rxVol += float64(rxSpeed) * 0.5 / 1024.0
			txVol += float64(txSpeed) * 0.5 / 1024.0
		}

		// Compile the delimited telemetry layout frame payload
		routesList := strings.Join(active, ",")
		payload := fmt.Sprintf("TELEMETRY|%.1f MB|%d KB/s|%.1f MB|%d KB/s|%s\n", rxVol, rxSpeed, txVol, txSpeed, routesList)
		
		_, err := conn.Write([]byte(payload))
		if err != nil {
			log.Println("[-] Dashboard detached or closed window. Session cleaned up cleanly.")
			break // 💥 Break out of this sub-loop to let the accepter cycle back natively!
		}
	}
}

