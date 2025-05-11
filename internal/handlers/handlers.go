// handlers.go
package handlers

import (
	// std go libs
	"context"      // for context
	"database/sql" // for sql errors
	"errors"       // for error handling
	"fmt"          // print errors
	"os"           // for file reading/writing
	"time"         // context timeout

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"      // for State and Command
	"github.com/PietPadda/aggregator/internal/database" // for DB Go code from SQLC
	"github.com/PietPadda/aggregator/internal/rssfeed"  // for RSS feed fetching
	"github.com/google/uuid"                            // for UUID generation
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

	// check if user already exists in database
	_, err := s.DB.GetUser(context.Background(), username)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user exists check
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("error: user '%s' doesn't exist\n", username)
		os.Exit(1) // clean exit
	}
	// errors.Is sql.ErrNoRows > err = sql.ErrNoRows
	// why? it includes wrapped errors, the error returned may not match exactly!

	// getuser db check
	if err != nil {
		return fmt.Errorf("error getting user from db: %w", err)
	}
	// this is a generic error check, if it isn't sql.ErrNoRows, then it's something else

	// use state to access config and set username
	err = s.Config.SetUser(username)
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

// register handler logic
// NOTE: cmd will be register, and state holds the config file to "register" new user if he doesn't exist
func HandlerRegister(s *app.State, cmd app.Command) error {
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

	// get user id as UUID and timestamp for created/updated at fields
	id := uuid.New()          // generate new UUID
	currentTime := time.Now() // get current time

	/* Note: the methods & struct that SQLC generated
	METHOD CreateUser:

	func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
		row := q.db.QueryRowContext(ctx, createUser,
			arg.ID,
			arg.CreatedAt,
			arg.UpdatedAt,
			arg.Name,
		)
		var i User
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
		)
		return i, err
	}

	METHOD GetUser:

	func (q *Queries) GetUser(ctx context.Context, name string) (User, error) {
		row := q.db.QueryRowContext(ctx, getUser, name)
		var i User
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
		)
		return i, err
	}

	STRUCT CreateUserParams:

	type CreateUserParams struct {
	    ID        uuid.UUID
	    CreatedAt time.Time
	    UpdatedAt time.Time
	    Name      string
	} */

	// check if user already exists in database
	_, err := s.DB.GetUser(context.Background(), username)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// general check that's NOT checking if users exists
	if err != nil && !errors.Is(err, sql.ErrNoRows) { // check if err is NOT sql.ErrNoRows
		return fmt.Errorf("error getting user from db: %w", err)
	}
	// this is a generic error check, if it isn't sql.ErrNoRows, then it's something else
	// errors.Is sql.ErrNoRows > err = sql.ErrNoRows
	// why? it includes wrapped errors, the error returned may not match exactly!

	// user exits check
	if err == nil {
		fmt.Printf("error: user '%s' exists\n", username)
		os.Exit(1) // clean exit
	}

	// user doesn't exist, so we can make a new user
	// create/register new user in database

	user, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        id,          // set id to UUID
		CreatedAt: currentTime, // set created at to current time
		UpdatedAt: currentTime, // set updated at to current time
		Name:      username,    // set name to username, arg of th ecommand
	})
	// CreateUser is a method from DB pass through state s (we made using users.sql)
	// CreateUserParams is a struct that was genned in database package
	// could do "_, err := ..." but we need user for printing confirmation msg
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user registration check
	if err != nil {
		return fmt.Errorf("error registering user: %w", err)
	}

	// use state to access config and set username
	err = s.Config.SetUser(username)
	// apply method to config file, which is contained in state

	// username set check
	if err != nil {
		return fmt.Errorf("error setting username: %w", err)
	}

	// print confirmation msg to user + log user details
	fmt.Printf("User '%s' has successfully been registered!\n", username)                     // confirmation msg
	fmt.Printf("User details:\n  ID = %s\n  CreatedAt = %s\n  UpdatedAt = %s\n  Name = %s\n", // log user details
		user.ID, user.CreatedAt, user.UpdatedAt, user.Name)

	// return success
	return nil
}

// reset handler logic
// NOTE: cmd will be reset, and state holds the config file to "reset" the users table
// NOTE: this is a dangerous command, so be careful with it! (for production code! but for our little app, it's fine)
func HandlerReset(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		fmt.Printf("error: State is nil")
		os.Exit(1) // clean exit code 1
	}

	/* Note: the method that SQLC generated
	METHOD Reset:

	func (q *Queries) Reset(ctx context.Context) error {
		_, err := q.db.ExecContext(ctx, reset)
		return err
	} */

	// run the reset command
	err := s.DB.Reset(context.Background())
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// reset check
	if err != nil {
		fmt.Printf("error resetting database: %s\n", err)
		os.Exit(1) // clean exit code 1
	}

	// success with code 0
	fmt.Printf("Database successfully reset!\n")
	os.Exit(0) // clean exit code 0

	// return success
	return nil
	// this will never be reached, but it's here for the requirement of the Register function
	// and to make the function complete
}

// getusers handler logic
// NOTE: cmd will be users, and state holds the config file to "users" from users table, also showing "current" user
func HandlerGetUsers(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		fmt.Printf("error: State is nil")
		os.Exit(1) // clean exit code 1
	}

	/* Note: the method that SQLC generated
		METHOD GetUsers:

	func (q *Queries) GetUsers(ctx context.Context) ([]string, error) {
	    rows, err := q.db.QueryContext(ctx, getUsers)
	    if err != nil {
	        return nil, err
	    }
	    defer rows.Close()
	    var items []string
	    for rows.Next() {
	        var name string
	        if err := rows.Scan(&name); err != nil {
	            return nil, err
	        }
	        items = append(items, name)
	    }
	    if err := rows.Close(); err != nil {
	        return nil, err
	    }
	    if err := rows.Err(); err != nil {
	        return nil, err
	    }
	    return items, nil
	} */

	// run the getusers command
	users, err := s.DB.GetUsers(context.Background())
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// getusers check
	if err != nil {
		fmt.Printf("error returning registered users from database: %s\n", err)
		os.Exit(1) // clean exit code 1
	}

	// no users check
	if len(users) == 0 {
		fmt.Printf("No users registered in database!\n")
		os.Exit(0) // clean exit code 0
	}

	// nil current user check
	if s.Config.Name == nil {
		fmt.Printf("error: current user is nil\n")
		os.Exit(1) // clean exit code 1
	}

	// get current user (safely deref after checking nil ptr)
	currentUser := *s.Config.Name // deref as it's *string :)

	// print users from database
	for _, user := range users {
		// check if current user
		if user == currentUser {
			fmt.Printf("* %s (current)\n", user)
			continue // skip to next user (else we print it twice)
		}
		fmt.Printf("* %s\n", user)
	}
	// succesfully printed users
	os.Exit(0) // clean exit code 0

	// return success
	return nil
	// this will never be reached, but it's here for the requirement of the Register function
	// and to make the function complete
}

// agg handler logic
// NOTE: cmd will be agg, and state holds the config file, it will "agg"regate the RSSFeed
// FetchFeed handles the error checking and parsing of the RSS feed
func HandlerAgg(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Don't forget to cancel to prevent resource leaks
	// context.WithTimeout creates a new context with a timeout of 10 seconds
	// this is used to limit the time the function can run, in case of a slow network or server
	// cancel is a function that cancels the context, and should be called when done

	// run the fetchfeed command
	feed, err := rssfeed.FetchFeed(ctx, "https://www.wagslane.dev/index.xml")

	// fetchfeed check
	if err != nil {
		return fmt.Errorf("error aggregating RSS feed: %w", err)
	}

	// print the entire RSSFeed struct
	// HEADER
	fmt.Printf("RSS Feed:\n")

	// CHANNEL
	fmt.Printf("  Title: %s\n", feed.Channel.Title)
	fmt.Printf("  Link: %s\n", feed.Channel.Link)
	fmt.Printf("  Description: %s\n", feed.Channel.Description)
	fmt.Printf("  Generator: %s\n", feed.Channel.Generator)
	fmt.Printf("  Language: %s\n", feed.Channel.Language)
	fmt.Printf("  LastBuildDate: %s\n", feed.Channel.LastBuildDate)
	fmt.Printf("  Atom: %s\n", feed.Channel.Atom.Href)
	fmt.Printf("  Atom Rel: %s\n", feed.Channel.Atom.Rel)
	fmt.Printf("  Atom Type: %s\n", feed.Channel.Atom.Type)

	// RSSFEED
	fmt.Printf("  Items:\n")
	for _, item := range feed.Channel.Items {
		fmt.Printf("    Title: %s\n", item.Title)
		fmt.Printf("    Link: %s\n", item.Link)
		fmt.Printf("    PubDate: %s\n", item.PubDate)
		fmt.Printf("    GUID: %s\n", item.GUID)
		fmt.Printf("    Description: %s\n", item.Description)
	}

	// succesfully printed RSSFeed user confirmation msg
	fmt.Printf("RSS Feed successfully aggregated!\n")

	// return success
	return nil
}
