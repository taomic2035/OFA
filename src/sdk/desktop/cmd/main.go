// OFA Desktop Agent - Main entry point
// Supports Windows, macOS, and Linux
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ofa/sdk/desktop"
)

var (
	version   = "2.0.0"
	buildDate = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	install := flag.Bool("install", false, "Install as system service")
	uninstall := flag.Bool("uninstall", false, "Uninstall system service")
	start := flag.Bool("start", false, "Start the agent")
	flag.Parse()

	if *showVersion {
		fmt.Printf("OFA Desktop Agent v%s (built %s)\n", version, buildDate)
		return
	}

	// Load or create configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config: %v, using defaults", err)
		config = desktop.DefaultAgentConfig()
	}

	// Handle service installation
	if *install {
		err := installService(config)
		if err != nil {
			log.Fatalf("Failed to install service: %v", err)
		}
		fmt.Println("Service installed successfully")
		return
	}

	if *uninstall {
		err := uninstallService(config)
		if err != nil {
			log.Fatalf("Failed to uninstall service: %v", err)
		}
		fmt.Println("Service uninstalled successfully")
		return
	}

	// Create agent
	agent, err := desktop.NewDesktopAgent(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		agent.Stop()
	}()

	// Start agent
	log.Printf("Starting OFA Desktop Agent v%s...", version)
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// Wait for completion
	select {}
}

// loadConfig loads configuration from file or creates default
func loadConfig(path string) (*desktop.AgentConfig, error) {
	if path == "" {
		// Use default config location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(homeDir, ".ofa", "agent", "config.json")
	}

	// Check if config exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		config := desktop.DefaultAgentConfig()
		config.ID = generateAgentID()
		config.Name = fmt.Sprintf("Desktop-%s", config.ID[:8])

		// Save config
		if err := saveConfig(path, config); err != nil {
			log.Printf("Warning: Failed to save default config: %v", err)
		}

		return config, nil
	}

	return desktop.LoadConfig(path)
}

// saveConfig saves configuration to file
func saveConfig(path string, config *desktop.AgentConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := os.Create(path)
	if err != nil {
		return err
	}
	defer data.Close()

	// Write config
	encoder := json.NewEncoder(data)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// generateAgentID generates a unique agent ID
func generateAgentID() string {
	return fmt.Sprintf("desktop-%d", time.Now().UnixNano())
}

// installService installs the agent as a system service
func installService(config *desktop.AgentConfig) error {
	// Platform-specific service installation
	return fmt.Errorf("service installation not implemented for this platform")
}

// uninstallService uninstalls the system service
func uninstallService(config *desktop.AgentConfig) error {
	return fmt.Errorf("service uninstallation not implemented for this platform")
}