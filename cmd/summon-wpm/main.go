package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/infamousjoeg/summon-wpm/internal/auth"
	"github.com/infamousjoeg/summon-wpm/internal/config"
	"github.com/infamousjoeg/summon-wpm/internal/provider"
)

const version = "0.1.0"

func main() {
	var showHelp, showVersion, configureFlag, loginFlag, verbose bool

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&configureFlag, "config", false, "Configure the provider")
	flag.BoolVar(&loginFlag, "login", false, "Login to CyberArk Identity")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")

	flag.Parse()

	if showVersion {
		fmt.Printf("summon-wpm v%s\n", version)
		os.Exit(0)
	}

	if showHelp {
		showUsage()
		os.Exit(0)
	}

	configFile := config.GetConfigFilePath()

	if configureFlag {
		config.RunConfigWizard(configFile)
		os.Exit(0)
	}

	if loginFlag {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error loading config: %s\n", err)
			}
			config.RunConfigWizard(configFile)
			cfg, err = config.LoadConfig(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %s\n", err)
				os.Exit(1)
			}
		}

		forceInteractive := !(cfg.ClientID != "" && cfg.ClientSecret != "")

		if err := auth.Authenticate(cfg, configFile, forceInteractive); err != nil {
			fmt.Fprintf(os.Stderr, "Authentication failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Authentication successful")
		os.Exit(0)
	}

	// Get the variable name from command line arguments
	args := flag.Args()
	if len(args) != 1 {
		showUsage()
		os.Exit(1)
	}

	appID := args[0]

	if verbose {
		fmt.Fprintf(os.Stderr, "Looking up app credentials for: %s\n", appID)
	}

	// Create the provider and execute it
	p := provider.NewProvider(verbose)
	result, err := p.GetCredential(appID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Success! Output the password to stdout
	fmt.Print(result)
}

func showUsage() {
	fmt.Println("CyberArk Workload Password Management Summon Provider")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  summon-wpm [options] <app_id>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("  --config       Run the configuration wizard")
	fmt.Println("  --login        Login to CyberArk Identity")
	fmt.Println("  --verbose      Enable verbose output")
	fmt.Println()
	fmt.Println("For use with Summon (https://github.com/cyberark/summon)")
}
