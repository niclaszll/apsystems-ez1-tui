package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/niclaszll/apsystems-ez1-tui/internal/tui"
	"github.com/niclaszll/apsystems-ez1-tui/pkg/apsystems"
)

// Set via ldflags during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	host := flag.String("host", "", "Microinverter IP address or hostname (required)")
	port := flag.Int("port", 8050, "Microinverter API port")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ez1-tui %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	if *host == "" {
		fmt.Println("Error: -host flag is required")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  ez1-tui -host 192.168.1.100")
		fmt.Println("  ez1-tui -host 192.168.1.100 -port 8050")
		os.Exit(1)
	}

	client := apsystems.NewClient(*host, *port)

	model := tui.NewModel(client)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
