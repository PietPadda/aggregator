// main.go
package main

import (
	// standard go libarries
	"fmt" // for printing
	"os"  // for file reading/writing

	// internal packages
	"github.com/PietPadda/aggregator/internal/config"
)

func main() {
	// read the config file
	_, err := config.Read()
	// _,. because we're only using it when printing to terminal!

	// read check
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1) // clean exit
	}

	// set user to lane
	_, err = config.SetUser("lane")

	// username set check
	if err != nil {
		fmt.Println("Error setting username:", err)
		os.Exit(1) // clean exit
	}

	// read the config file again
	cfg, err := config.Read()

	// read check (again)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1) // clean exit
	}

	// print config contents to the terminal
	fmt.Println("Config file contents: ", cfg)
}
