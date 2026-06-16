package main

import (
	"bufio"
	"fmt"
	"image/color" // 🛠️ THE FIX: Native color interface package imported!
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

        "charm.land/lipgloss/v2"
        "dumb-proxy-go/pkg/termutil"
)

// =========================================================================
// 🎨 THE UNIVERSAL PALETTE STRUCT (NATIVE INTEGRATION)
// =========================================================================
type Palette struct {
	Background color.Color // 🛠️ THE FIX: Clean decoupled color.Color interface fields!
	Primary    color.Color
	Secondary    color.Color
	Text       color.Color
	Success    color.Color
	Fail       color.Color
}

var JungleTheme = Palette{
	Background: lipgloss.Color("#141714"),
	Primary:    lipgloss.Color("#414833"),
	Secondary:    lipgloss.Color("#A68A64"),
	Text:       lipgloss.Color("#EDE0D4"),
	Success:    lipgloss.Color("#7AE582"),
	Fail:       lipgloss.Color("#E63946"),
}

// Live Global Shared Metrics States
var (
	udsPath       = "/tmp/dumb-proxy.sock"
	activeStreams = []string{}
	
	rxVolume = "0.0 MB"
	rxSpeed  = "0 KB/s"
	txVolume = "0.0 MB"
	txSpeed  = "0 KB/s"

	spinnerFrames = []string{"💥", "🔥", "✨", "🍌", "⚡", "🌟", "🎉"}
	spinnerIndex  int
	startTime     = time.Now()
)

// =========================================================================
// 🎛️ COMPONENT SEPARATION OF CONCERNS (CLEAN DESIGN EXTRACTIONS)
// =========================================================================

func renderHeader(cols int) string {
	// The universal baseline structure
	base := lipgloss.NewStyle().
		Bold(true).
		Foreground(JungleTheme.Text).
		Background(JungleTheme.Primary)

	// Build the continuous row segments safely using native string concatenations!
	text := base.Render("🦍 DUMB-PROXY-GO // DAEMON: [ ") +
		base.Foreground(JungleTheme.Success).Render("ONLINE") +
		base.Render(" ] // SYSTEM PROXY: [ ") +
		base.Foreground(JungleTheme.Success).Render("ACTIVE") +
		base.Render(" ]")

	return lipgloss.NewStyle().
		Width(cols).
		Align(lipgloss.Center).
		Background(JungleTheme.Primary).
		Render(text)
}

func renderPerformanceBox(panelWidth int) string {
	label := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Text)
	secondary  := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Secondary)

	uptime := func() string {
		uptimeStr := time.Since(startTime).Round(time.Second).String()
		return label.Render("DAEMON UPTIME: " + uptimeStr)
	}

	rx := func() string {
		header := label.Render("INGESTION DATA PIPELINE [RX]")
		stats  := fmt.Sprintf("Volume: %s | Speed: %s", label.Render(rxVolume), secondary.Render(rxSpeed))
		bar    := secondary.Render("[===========>         ]")
		return lipgloss.JoinVertical(lipgloss.Left, header, stats, bar)
	}

	tx := func() string {
		header := label.Render("TRANSMISSION RETRY PIPELINE [TX]")
		stats  := fmt.Sprintf("Volume: %s | Speed: %s", label.Render(txVolume), secondary.Render(txSpeed))
		bar    := secondary.Render("[==>                  ]")
		return lipgloss.JoinVertical(lipgloss.Left, header, stats, bar)
	}

	return lipgloss.NewStyle().
		Width(panelWidth).
		PaddingLeft(4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			uptime(),
			"",
			rx(),
			"",
			tx(),
		))
}

func renderRouteBox(panelWidth int) string {
	// 🛠️ THE ATOM BASES: Core typography styles pulling straight from our theme struct
	label     := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Text)
	secondary := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Secondary)
	success   := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Success)
	divider   := lipgloss.NewStyle().Foreground(JungleTheme.Primary)

	// --- 1. HEADER & SPLIT LINE SUB-COMPONENT LAMBDA ---
	header := func() string {
		titleLine   := label.Render("MULTIPLEX ROUTE REGISTRY")
		dividerLine := divider.Render(strings.Repeat("-", panelWidth-4))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, dividerLine)
	}

	// --- 2. MULTIPLEXED TRACKER LINES SUB-COMPONENT LAMBDA ---
	routes := func() string {
		if len(activeStreams) == 0 {
			return "" // Return nothing so the container skips space if empty
		}

		var rows []string
		for _, target := range activeStreams {
			// Prevent line wrapping breakage if the terminal column window gets too narrow
			if len(target) > panelWidth-8 {
				target = target[:panelWidth-11] + "..."
			}
			// Sharp vector arrow tinted with secondary khaki accent for perfect color symmetry
			rows = append(rows, fmt.Sprintf("%s » %s", label.Render("[🌐] PIPE ROUTE"), success.Render(target)))
		}
		return strings.Join(rows, "\n")
	}

	// --- 3. DYNAMIC STATUS ENGINE / SPIN LOG SUB-COMPONENT LAMBDA ---
	shuffle := func() string {
		if len(activeStreams) == 0 {
			// Using the secondary palette color to keep the chill zone text soft
			return secondary.Render("(No active traffic. Engine chilling... 🦍💤)")
		}
		
		currentSpin := spinnerFrames[spinnerIndex%len(spinnerFrames)]
		return label.Render("🕺 GORILLA SHUFFLE ENGAGED ") + 
			success.Render(currentSpin) + 
			label.Render(" 🌪️")
	}

	// 🛠️ THE MASTER CONTAINER JOIN:
	// Stacks Header, Routes list, and the Dance Shuffle components natively with clean vertical padding!
	return lipgloss.NewStyle().
		Width(panelWidth).
		PaddingLeft(4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			header(),
			"", // Clean native spacing line
			routes(),
			"",
			shuffle(),
		))
}

// =========================================================================
// 🚀 MAIN LAYOUT ENGINE OVERSEER
// =========================================================================

func drawOperationsCenter() {
	cols, _ := termutil.GetSize()
	fmt.Print("\033[H\033[2J\033[?25l") // Hard clear canvas screen buffer

	// Establish structural layout width boundaries
	panelWidth := (cols / 2) - 2
	if panelWidth < 20 {
		panelWidth = 20
	}

	// 🛠️ THE ATOM BASES: Clean styling references pulling straight from our theme struct
	divider   := lipgloss.NewStyle().Foreground(JungleTheme.Primary)
	tabStyle  := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Text).PaddingLeft(2)
	secAccent := lipgloss.NewStyle().Bold(true).Foreground(JungleTheme.Secondary)

	// --- 1. VIEW SELECTOR TAB BAR SUB-COMPONENT LAMBDA ---
	tabs := func() string {
		// Cleanly highlight our active screen tab utilizing our Secondary Khaki Accent! 🌿
		activeTab  := secAccent.Render("MAIN APP CONTROL CENTER")
		inactiveTab := lipgloss.NewStyle().Foreground(JungleTheme.Primary).Render("SYSTEM OPTIONS COCKPIT")
		return tabStyle.Render(fmt.Sprintf("%s  |  %s", activeTab, inactiveTab))
	}

	// --- 2. SIDE-BY-SIDE SIDE PANEL HORIZONTAL ASSEMBLER LAMBDA ---
	grid := func() string {
		leftBox  := renderPerformanceBox(panelWidth)
		rightBox := renderRouteBox(panelWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
	}

	// --- 3. VIM COMMAND INPUT FOOTER BAR SUB-COMPONENT LAMBDA ---
	footer := func() string {
		cursor := lipgloss.NewStyle().Foreground(JungleTheme.Success).Blink(true).Render("_")
		return lipgloss.NewStyle().Foreground(JungleTheme.Text).Render(":" + cursor)
	}

	// 🛠️ THE MASTER SCREEN CONTAINER JOIN:
	// Stacks Header, Tabs, Lines, Panels Grid, and Input Footer in 100% pure vertical vertical alignment!
	dividerLine := divider.Render(strings.Repeat("-", cols))
	
	fmt.Println(lipgloss.JoinVertical(
		lipgloss.Left,
		renderHeader(cols),
		tabs(),
		dividerLine,
		grid(),
		"", // Clean native empty line row padding before input
		dividerLine,
		footer(),
	))
}

func main() {
	conn, err := net.Dial("unix", udsPath)
	if err != nil {
		log.Fatalf("[-] Tool cannot attach. Boot mock_daemon.go first!: %v", err)
	}
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print("\033[?25h\033[0m\n")
		os.Exit(0)
	}()

	// 🛠️ THE SINGLE MASTER RENDER LOOP TICKER (LOCKED AT 100MS TO STOP TEARING)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			spinnerIndex++
			drawOperationsCenter() // The ONLY function allowed to write to stdout!
		}
	}()

	// Background network socket reader updates states silently in memory
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "|")
		
		if len(parts) >= 6 && parts[0] == "TELEMETRY" {
			// 🛠️ NO DRAW CALLS HERE! Only update the variables smoothly.
			rxVolume = parts[1]
			rxSpeed = parts[2]
			txVolume = parts[3]
			txSpeed = parts[4]
			
			if parts[5] == "" {
				activeStreams = []string{}
			} else {
				activeStreams = strings.Split(parts[5], ",")
				sort.Strings(activeStreams)
			}
		}
	}
}

