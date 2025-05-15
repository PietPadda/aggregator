// handlers.go
package handlers

import (
	// std go libs
	"context"      // for context
	"database/sql" // for sql errors
	"errors"       // for error handling
	"fmt"          // print errors
	"os"           // for file reading/writing
	"strings"      // filter text in strs
	"time"         // context timeout

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"      // for State and Command
	"github.com/PietPadda/aggregator/internal/database" // for DB Go code from SQLC
	"github.com/PietPadda/aggregator/internal/rssfeed"  // for RSS feed fetching
	"github.com/google/uuid"                            // for UUID generation
)

// MIDDLEWARE

// used to replace GetUser() and error checking inside handler command functions
// PascalCase to export to main.go (as opposed to camelCase for package only)
// NOTE: database.User --> database package has file models.go with User struct
func MiddlewareLoggedIn(handler func(s *app.State, cmd app.Command, user database.User) error) func(*app.State, app.Command) error {
	return func(s *app.State, cmd app.Command) error {
		// state check
		if s == nil {
			return fmt.Errorf("error: State is nil")
		}

		// config check
		if s.Config == nil {
			return fmt.Errorf("error: Config is nil")

		}

		// logged in user check (allow safe dereffing)
		if s.Config.Name == nil {
			return fmt.Errorf("error: user is not logged in")
		}

		// empty username check
		if *s.Config.Name == "" {
			return fmt.Errorf("error: username is empty")
		} //*s because we used *string!
		// currentUser := *s.Config.Name

		// get current user struct
		user, err := s.DB.GetUser(context.Background(), *s.Config.Name)

		// getuser check
		if err != nil {
			return fmt.Errorf("error getting user from db: %w", err)
		}

		// return the handler command from Handler map in config, in state
		return handler(s, cmd, user)
		// NOTE: we choose the name as "handler"
	}
}

// COMMAND HANDLERS

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

	// cmd input check
	// command is a struct, get its field for length check
	if len(cmd.Args) < 1 {
		return fmt.Errorf("error: time between requests required")
	} // agg handler expects ONE arg: time_between_reqs!!

	// get arguments input
	timeInput := cmd.Args[0] // not needed, but nicely readable!

	// parse the time duration string
	timeBetweenRequests, err := time.ParseDuration(timeInput)

	// time between requests check
	if err != nil {
		return fmt.Errorf("error: invalid duration format: %w", err)
	}

	// init the infinite loop time.Ticker()
	timeTicker := time.NewTicker(timeBetweenRequests)
	defer timeTicker.Stop() // stop the ticker when command ends
	// CORE: it will NEVER end because it's an infinite loop!
	// but good practice to add in anyway

	// inform user of the time interval
	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	// start an infinite loop with a time.Ticker()
	for {
		// scrape the feeds immediately!
		err = scrapeFeeds(s.DB)

		// scrape feeds check
		if err != nil {
			fmt.Printf("error scraping the feeds: %s\n", err)
		}

		// block the loop and wait until ticker TICKS!
		<-timeTicker.C // ticker runs on it's own channel called C
		// <- = reads from the ticker channel, which sends the current time at regular intervals
		// we set the "tick" to be timeBetweenRequests
	}

	// return success
	return nil
	// this UNREACHABLE code, as we have an infinite loop that only ends on ctrl+c!!!!
	// just formality as function returns an error value
}

// addfeed handler logic
// NOTE: cmd will be addfeed, and state holds the config file, it will add a new feed to the database
// now use middleware to provide user as input! not more GetUser()!
func HandlerAddFeed(s *app.State, cmd app.Command, user database.User) error {
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

	// THIS PART IS NOW HANDLED BY MIDDLEWARELOGIN!
	// get user by currentUser from database to set the feed follow's fk user_id
	// user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

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
	// we'll use our "hacky" followCmd command to do this!
	err = HandlerFollow(s, followCmd, user)
	// NOTE: added user as it's required for our middleware :)

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
// now use middleware to provide user as input! not more GetUser()!
func HandlerFollow(s *app.State, cmd app.Command, user database.User) error {
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

	// THIS PART IS NOW HANDLED BY MIDDLEWARELOGIN!
	// get user by currentUser from database to set the feed follow's fk user_id
	// user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

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
// now use middleware to provide user as input! not more GetUser()!
func HandlerFollowing(s *app.State, cmd app.Command, user database.User) error {
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

	// get current user safely from MIDDELWARE!
	currentUser := user.Name

	// THIS PART IS NOW HANDLED BY MIDDLEWARELOGIN!
	// get user by currentUser from database to set the feed follow's fk user_id
	// user, err := s.DB.GetUser(context.Background(), currentUser)
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

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

// unfollow handler logic
// NOTE: cmd will be unfollow, and state holds the config file to "unfollow" a feed follow from current user
func HandlerUnfollow(s *app.State, cmd app.Command, user database.User) error {
	// state ptr check
	if s == nil {
		fmt.Printf("error: State is nil")
		os.Exit(1) // clean exit code 1
	}

	/* Note: the method and struct that SQLC generated
	METHOD DeleteFeedFollowByUserAndFeed:

	func (q *Queries) DeleteFeedFollowByUserAndFeed(ctx context.Context, arg DeleteFeedFollowByUserAndFeedParams) (FeedFollow, error) {
		row := q.db.QueryRowContext(ctx, deleteFeedFollowByUserAndFeed, arg.Url, arg.UserID)
		var i FeedFollow
		err := row.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
			&i.FeedID,
		)
		return i, err
	}


	STRUCTD DeleteFeedFollowByUserAndFeedParams:

	type DeleteFeedFollowByUserAndFeedParams struct {
		Url    string
		UserID uuid.UUID
	} */

	// nil current user check
	if s.Config.Name == nil {
		return fmt.Errorf("error: current user is nil/not logged in")
	}

	// cmd input check
	// command is a struct, get its field for length check
	if len(cmd.Args) < 1 {
		return fmt.Errorf("error: feed url required")
	} // login handler expects TWO arg: feed NAME and URL!

	// get arguments input
	feedURL := cmd.Args[0] // not needed, but nicely readable!

	// get current user safely from MIDDLEWARE!
	currentUserID := user.ID
	currentUser := user.Name

	// create SQL struct

	// run the unfollow command
	_, err := s.DB.DeleteFeedFollowByUserAndFeed(context.Background(), database.DeleteFeedFollowByUserAndFeedParams{
		Url:    feedURL,       // set feed url from arg
		UserID: currentUserID, // set user id from middleware
	})
	// DeleteFeedFollowByUserAndFeed is a method from DB pass through state s (we made using feed_follows.sql)
	// DeleteFeedFollowByUserAndFeedParams is a struct that was genned in database package
	// doing "_, err := ..." because we're not using the feed follow!
	// context.Background() provides root empty context with no deadlines or cancellation - required by DB API

	// feed follow exists check
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("%s is not following this feed!\n", currentUser)
		os.Exit(0) // clean exit
	}
	// errors.Is sql.ErrNoRows > err = sql.ErrNoRows
	// why? it includes wrapped errors, the error returned may not match exactly!

	// unfollow check
	if err != nil {
		fmt.Printf("error unfollowing feed: %s\n", err)
		os.Exit(1) // clean exit code 1
	}

	// success with code 0
	fmt.Printf("Feed successfully unfollowed!\n")
	os.Exit(0) // clean exit code 0

	// return success
	return nil
	// this will never be reached, but it's here for the requirement of the Register function
	// and to make the function complete
}

// HELPER FUNCTIONS

// aggregation function helper to get feeds for agg command
func scrapeFeeds(queries *database.Queries) error {
	// database queries check
	if queries == nil {
		return fmt.Errorf("error: database queries is nil")
	}

	// use GetNextFeedToFetch to... get the next feed to fetch!
	nextFeed, err := queries.GetNextFeedToFetch(context.Background())

	// get next feed to fetch check
	if err != nil {
		// check if no needs available in DB
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("No feeds found in database. Add some using the 'addfeed' command.")
			return nil
		}
		return fmt.Errorf("error getting next feed to fetch: %w", err)
	}

	// get feed data (not needed, but nice and readable!)
	feedID := nextFeed.ID
	feedName := nextFeed.Name
	feedURL := nextFeed.Url

	// mark the feed as fetched
	err = queries.MarkFeedFetched(context.Background(), feedID)

	// mark feed fetched check
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	// tell user that fetching has started!
	fmt.Printf("Fetching feed: %s (%s)\n", feedName, feedURL)

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Don't forget to cancel to prevent resource leaks
	// context.WithTimeout creates a new context with a timeout of 10 seconds
	// this is used to limit the time the function can run, in case of a slow network or server
	// cancel is a function that cancels the context, and should be called when done

	// fetch the feed using url (from rssfeed.go)
	feed, err := rssfeed.FetchFeed(ctx, feedURL)

	// fetch feed check
	if err != nil {
		return fmt.Errorf("error fetching the marked feed %s: %w", feedName, err)
	}

	// print the feed info
	fmt.Printf("Feed: %s\n", feedName)

	// loop over rssfeed and print title of each item in feed
	for _, item := range feed.Channel.Items {
		fmt.Printf(" - %s\n", item.Title)
	}

	// print newline for visual clairty
	fmt.Println()

	// return success
	return nil
}
