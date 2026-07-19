package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "dev"

func main() {
	// Parse command line flags
	helpFlag := flag.Bool("help", false, "Show help message")
	hFlag := flag.Bool("h", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show help if requested
	if *helpFlag || *hFlag {
		showHelp()
		return
	}

	// Show version if requested
	if *versionFlag {
		fmt.Printf("Clip version %s\n", version)
		return
	}

	// Default behavior: show help if no arguments provided
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	// For now, just show help for any other arguments
	showHelp()
}

func showHelp() {
	fmt.Println("Clip")
	fmt.Println()
	fmt.Println("A fast CLI for extracting text from documents.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  clip [arguments]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --help, -h    Show help message")
	fmt.Println("  --version     Show version information")
}