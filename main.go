// main.go
package main

import (
	// standard go libarries
	"fmt" // for printing
	"os"  // for file reading/writing

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"
	"github.com/PietPadda/aggregator/internal/config"
	"github.com/PietPadda/aggregator/internal/handlers"
)

func main() {
	// read the config file
	cfg, err := config.Read()
	// _,. because we're only using it when printing to terminal!
	// UPDATE: also for SetUser to work as a method!

	// read check
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1) // clean exit
	}

	// create state instance and store config in
	state := &app.State{ // app
		Config: &cfg,
	}
	// we declare state as a ptr to app.State, thus use &app! (our funcs use s *State !)
	// Config is ptr in the State struct, thus we use &cfg

	// create commands instance with init map of handler functions
	cmds := &app.Commands{
		Handler: make(map[string]func(*app.State, app.Command) error), // matches the struct
	}
	// we declare commands as a ptr to app.Commands, thus use &app! (our funcs use c *Commands !)
	// Handler is in Commands struct, and we have to init the map! takes State ptr and Command!
	// why init the map? Because Go maps need to be init before they can be used! prevents Go panic

	// register the  handler function for the login cmd
	cmds.Register("login", handlers.HandlerLogin)
	// Registers receivces commands
	// "login" = the command we register
	// HandlerLogin works on handlers, and registers "login" there

	// CLI args check
	// 2 CLI args min! 1st = command, 2nd = arg
	if len(os.Args) < 2 {
		fmt.Println("error: insufficient arguments!")
		fmt.Println("Usage: aggregator <command> [args...]")
		os.Exit(1) // clean exit
	}

	// get args (not needed, readability!)
	cmdName := os.Args[1]  // aggregator <command>, thus index 1
	cmdArgs := os.Args[2:] // aggregator <command> [args...], thus index 2+

	// use CLI args to create a command
	cmd := app.Command{
		Name: cmdName, // command's name
		Args: cmdArgs, // args to the command
	}

	// run the command
	err = cmds.Run(state, cmd) // we created state, cmd and cmds above

	// run check
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1) // clean exit
	}
}
