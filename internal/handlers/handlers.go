// handlers.go
package handlers

import (
	// std go libs
	"context"
	"database/sql"
	"errors"
	"fmt" // print errors
	"os"
	"time"

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"
	"github.com/PietPadda/aggregator/internal/database"
	"github.com/google/uuid"
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
