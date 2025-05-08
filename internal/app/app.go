// app.gp
package app

import (
	// std go libraries
	"fmt" // printing errors

	// internal packages
	"github.com/PietPadda/aggregator/internal/config"
)

// app state struct
type State struct {
	Config *config.Config // ptr Config, Config type from config package
}

// cli command struct
type Command struct {
	Name string   // command name called in CLI
	Args []string // args passed to command
}

// commands handler struct
type Commands struct {
	Handler map[string]func(s *State, cmd Command) error // cmd map of key strs, takes state and cmd input
}

// register new command method
func (c *Commands) Register(name string, f func(*State, Command) error) error {
	// nil ptr check
	if c == nil {
		return fmt.Errorf("commands is nil")
	}

	// handler map init check
	if c.Handler == nil {
		return fmt.Errorf("Handler map is nil")
	}

	// register new command handler
	c.Handler[name] = f // register func f as key "name" to the Handler map in commands (c)

	// return success
	return nil
}

// run a registered cmd method
func (c *Commands) Run(s *State, cmd Command) error {
	// nil ptr check
	if c == nil {
		return fmt.Errorf("error: commands is nil")
	}

	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	// handler map init check
	if c.Handler == nil {
		return fmt.Errorf("error: Handler map is nil")
	}

	// get command name
	commandName := cmd.Name // not needed, but helps with readability

	// lookup command in Handler map (comma-ok)
	handler, ok := c.Handler[commandName] // takes a cmd name str as key

	// exist check
	if !ok {
		return fmt.Errorf("error: command is not registered: %s", commandName)
	}

	// return handler (which pass through an error)
	return handler(s, cmd)
	// we chose handler as name, and pass state and command, per func signature
}
