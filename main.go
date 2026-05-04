package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"phonecheck/checker"
)

func main() {
	typ := flag.String("t", "bitmap", "checker type: bitmap, bloom, sqlite, tree")
	dataFile := flag.String("f", "phone_numbers.bin", "data file path")
	flag.Parse()

	checker, err := checker.NewPhoneChecker(*typ, *dataFile)
	if err != nil {
		fmt.Printf("Failed to initialize checker: %v\n", err)
		os.Exit(1)
	}
	defer checker.Close()

	fmt.Println("Phone Checker initialized successfully")
	fmt.Println("Type phone numbers to check (type 'exit' to quit)")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "exit" || input == "quit" {
			break
		}
		if input == "" {
			continue
		}

		exists := checker.PhoneExists(input)
		if exists {
			fmt.Printf("✓ %s exists\n", input)
		} else {
			fmt.Printf("✗ %s does not exist\n", input)
		}
	}
}
