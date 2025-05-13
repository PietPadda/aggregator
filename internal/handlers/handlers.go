// handlers.go
package handlers

import (
	// std go libs
	"context"      // for context
	"database/sql" // for sql errors
	"errors"       // for error handling
	"fmt"          // print errors
	"os"           // for file reading/writing
	"strings"
	"time" // context timeout

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

// addfeed handler logic
// NOTE: cmd will be addfeed, and state holds the config file, it will add a new feed to the database
func HandlerAddFeed(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	// cmd input check
	// command is a struct, get its field for length check
	if len(cmd.Args) < 2 {
		return fmt.Errorf("error: feed name and url args required")
	} // login handler expects TWO arg: feed NAME and URL!

	// get arguments input
	feedName := cmd.Args[0] // not needed, but nicely readable!
	feedURL := cmd.Args[1]  // not needed, but nicely readable!

	// nil current user check
	if s.Config.Name == nil {
		return fmt.Errorf("error: current user is nil/not logged in")
	}

	// get current user (safely deref after checking nil ptr)
	currentUser := *s.Config.Name // deref as it's *string :)

	// get user by currentUser from database to set the feed's fk user_id
	user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user check
	if err != nil {
		return fmt.Errorf("error getting user from db: %w", err)
	}

	// get feed id as UUID and timestamp for created/updated at fields
	id := uuid.New()          // generate new UUID
	currentTime := time.Now() // get current time
	userID := user.ID         // get user id from user struct

	/* Note: the method & struct that SQLC generated
		METHOD GetUser:

	func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
		row := q.db.QueryRowContext(ctx, createFeed,
			arg.ID,
			arg.CreatedAt,
			arg.UpdatedAt,
			arg.Name,
			arg.Url,
			arg.UserID,
		)
		var i Feed
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
		)
		return i, err
	}

	STRUCT CreateFeedParams:
	type CreateFeedParams struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		Url       string
		UserID    uuid.UUID
	} */

	// create new feed in database
	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        id,          // set id to UUID
		CreatedAt: currentTime, // set created at to current time
		UpdatedAt: currentTime, // set updated at to current time
		Name:      feedName,    // set name to feedname, arg 0
		Url:       feedURL,     // set url to feedURL, arg 1
		UserID:    userID,      // set user id to current user
	})
	// CreateFeed is a method from DB pass through state s (we made using users.sql)
	// CreateFeedParams is a struct that was genned in database package
	// could do "_, err := ..." but we need feed for printing confirmation msg
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user registration check
	if err != nil {
		return fmt.Errorf("error adding new feed to database: %w", err)
	}

	// print confirmation msg to user + log feed details
	fmt.Printf("RSS Feed '%s' has successfully been added to database!\n", feedName)                                    // confirmation msg
	fmt.Printf("Feed details:\n  ID = %s\n  CreatedAt = %s\n  UpdatedAt = %s\n  Name = %s\n  URL: %s\n  UserID = %s\n", // log user details
		feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID)

	// before finishing, we also follow the feed for the current user :)
	// we COULD do all the logic, but HandlerFollow already exists
	// GO: function order doesn't matter in go, as long it's in the same package, we can just call it!

	// first let's create a new command for the follow handler
	// NOTE: this is a bit of a hack, but it works!
	followCmd := app.Command{
		Name: "follow",
		Args: []string{feedURL}, // we need the URL to follow the feed]
	}

	// Call the follow handler to follow the feed!
	// we'll use our "hacky" followCmd command to do thid!
	err = HandlerFollow(s, followCmd)

	// follow handler check
	if err != nil {
		// nice to have but NOT critical, so we just print a warning INSTEAD OF EXITING!
		fmt.Printf("Warning: error following feed: %s", err)
	}

	// return success
	return nil
}

// feeds handler logic
// NOTE: cmd will be feeds, and state holds the config file to db, feedname, feedurl and username that CREATED it
func HandlerFeeds(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		fmt.Printf("error: State is nil")
		os.Exit(1) // clean exit code 1
	}

	/* Note: the method that SQLC generated
	METHOD ListFeedsWithCreator:

	func (q *Queries) ListFeedsWithCreator(ctx context.Context) ([]ListFeedsWithCreatorRow, error) {
		rows, err := q.db.QueryContext(ctx, listFeedsWithCreator)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []ListFeedsWithCreatorRow
		for rows.Next() {
			var i ListFeedsWithCreatorRow
			if err := rows.Scan(&i.Feedname, &i.Feedurl, &i.Username); err != nil {
				return nil, err
			}
			items = append(items, i)
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return items, nil
	}

	STRUCT ListFeedsWithCreatorRow:

	type ListFeedsWithCreatorRow struct {
		Feedname string
		Feedurl  string
		Username string
	}*/

	// run the listfeedswithcreator sql query
	feeds, err := s.DB.ListFeedsWithCreator(context.Background())
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// listfeed check
	if err != nil {
		fmt.Printf("error returning feeds from database: %s\n", err)
		os.Exit(1) // clean exit code 1
	}

	// no feeds check
	if len(feeds) == 0 {
		fmt.Printf("No feeds logged in database!\n")
		os.Exit(0) // clean exit code 0
	}

	// print feeds header
	fmt.Println("Feeds list based on creator:")
	fmt.Println() // newline

	// print feeds from database
	for _, feed := range feeds {
		fmt.Printf("Feed name: %s\n", feed.Feedname)
		fmt.Printf("Feed URL: %s\n", feed.Feedurl)
		fmt.Printf("Created by: %s\n", feed.Username)
		fmt.Println() // newline
	}
	// succesfully printed users
	os.Exit(0) // clean exit code 0

	// return success
	return nil
	// this will never be reached, but it's here for the requirement of the Register function
	// and to make the function complete
}

// follow handler logic
// NOTE: cmd will be follow, and state holds the config file to allow current user to "follow" a feed
// essentially adds a feed follow record to the database
func HandlerFollow(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	// cmd input check
	// command is a struct, get its field for length check
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error: no command input")
	} // login handler expects ONE arg: the url!

	// get url input (first arg!)
	urlArg := cmd.Args[0] // not needed, but nicely readable!

	// get feed follow id as UUID and timestamp for created/updated at fields
	id := uuid.New()          // generate new UUID
	currentTime := time.Now() // get current time

	/* Note: the methods & struct that SQLC generated
	METHOD GetFeedByURL:

	func (q *Queries) GetFeedByURL(ctx context.Context, url string) (Feed, error) {
		row := q.db.QueryRowContext(ctx, getFeedByURL, url)
		var i Feed
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
		)
		return i, err
	}

	METHOD CreateFeedFollows:

	func (q *Queries) CreateFeedFollows(ctx context.Context, arg CreateFeedFollowsParams) (CreateFeedFollowsRow, error) {
		row := q.db.QueryRowContext(ctx, createFeedFollows,
			arg.ID,
			arg.CreatedAt,
			arg.UpdatedAt,
			arg.UserID,
			arg.FeedID,
		)
		var i CreateFeedFollowsRow
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
			&i.FeedID,
			&i.Username,
			&i.Feedname,
		)
		return i, err
	}

	STRUCT CreateFeedFollowsParams:

	type CreateFeedFollowsParams struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
		UserID    uuid.UUID
		FeedID    uuid.UUID
	} */

	// nil current user check
	if s.Config.Name == nil {
		return fmt.Errorf("error: current user is nil/not logged in")
	}

	// get current user (safely deref after checking nil ptr)
	currentUser := *s.Config.Name // deref as it's *string :)

	// get user by currentUser from database to set the feed follow's fk user_id
	user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user check
	if err != nil {
		return fmt.Errorf("error getting user from db: %w", err)
	}

	// get feed by url from database to set the feed follow's fk feed_id
	feed, err := s.DB.GetFeedByURL(context.Background(), urlArg)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user check
	if err != nil {
		return fmt.Errorf("error getting feed from db: %w", err)
	}

	// create new feed follow in database

	feedFollow, err := s.DB.CreateFeedFollows(context.Background(), database.CreateFeedFollowsParams{
		ID:        id,          // set id to UUID
		CreatedAt: currentTime, // set created at to current time
		UpdatedAt: currentTime, // set updated at to current time
		UserID:    user.ID,     // set user id to current user
		FeedID:    feed.ID,     // set feed id to feed by url
	})
	// CreateFeedFollows is a method from DB pass through state s (we made using users.sql)
	// CreateFeedFollowsParams is a struct that was genned in database package
	// could do "_, err := ..." but we need feed for printing confirmation msg
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// feed follow check
	if err != nil {
		// check if unique
		if strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("error: you are already following this feed")
		}
		// other general error
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	// print confirmation msg to user + log user details
	fmt.Printf("%s is now following: %s\n", // confirmation msg
		feedFollow.Username, feedFollow.Feedname)

	// return success
	return nil
}

// following handler logic
// NOTE: cmd will be following, and state holds the config file to print all feeds the current user is "following"
func HandlerFollowing(s *app.State, cmd app.Command) error {
	// state ptr check
	if s == nil {
		return fmt.Errorf("error: State is nil")
	}

	/* Note: the methods & struct that SQLC generated
	METHOD GetFeedFollowsForUser:

	func (q *Queries) GetFeedFollowsForUser(ctx context.Context, userID uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
		rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []GetFeedFollowsForUserRow
		for rows.Next() {
			var i GetFeedFollowsForUserRow
			if err := rows.Scan(
				&i.ID,
				&i.CreatedAt,
				&i.UpdatedAt,
				&i.UserID,
				&i.FeedID,
				&i.Username,
				&i.Feedname,
			); err != nil {
				return nil, err
			}
			items = append(items, i)
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return items, nil
	} */

	// nil current user check
	if s.Config.Name == nil {
		return fmt.Errorf("error: current user is nil/not logged in")
	}

	// get current user (safely deref after checking nil ptr)
	currentUser := *s.Config.Name // deref as it's *string :)

	// get user by currentUser from database to set the feed follow's fk user_id
	user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// user check
	if err != nil {
		return fmt.Errorf("error getting user from db: %w", err)
	}

	// run the getfeedfollowsfor user command
	feedFollows, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// getfeedfollowsofruser check
	if err != nil {
		fmt.Printf("error returning feed follows from database: %s\n", err)
		os.Exit(1) // clean exit code 1
	}

	// no feed follows check
	if len(feedFollows) == 0 {
		fmt.Printf("No feed follows in database!\n")
		os.Exit(0) // clean exit code 0
	}

	// print feeds follows header
	fmt.Printf("Feeds followed by %s:\n", currentUser)
	fmt.Println() // newline

	// print names of feed follows from database for current user
	for _, feedFollow := range feedFollows {
		fmt.Printf("Feed name: %s\n", feedFollow.Feedname)
		fmt.Println() // newline
	}
	// succesfully printed users
	os.Exit(0) // clean exit code 0

	// return success
	return nil
	// this will never be reached, but it's here for the requirement of the Register function
	// and to make the function complete
}
