// handlers.go
package handlers

import (
	// std go libs
	"fmt" // print errors

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"
)

// login handler logic
// NOTE: cmd will be login, and state holds the config file to "sign in" the user
func HandlerLogin(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	// cmd input check
	// command is a struct, get its field for length check
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error: no command input")
	} // login handler expects ONE arg: the username!

	// get username input (first arg!)
	username := cmd.Args[0] // not needed, but nicely readable!

	// use state to access config and set username
	err := s.Config.SetUser(username)
	// apply method to config file, which is contained in state

	// username set check
	if err != nil {
		return fmt.Errorf("error setting username: %w", err)
	}

	// print confirmation msg to user
	fmt.Printf("User '%s' has successfully logged in!\n", username)

	// return success
	return nil
}
