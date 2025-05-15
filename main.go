// main.go
package main

import (
	// standard go libarries
	"database/sql"
	"fmt" // for printing
	"os"  // for file reading/writing

	// internal packages
	"github.com/PietPadda/aggregator/internal/app"
	"github.com/PietPadda/aggregator/internal/config"
	"github.com/PietPadda/aggregator/internal/database"
	"github.com/PietPadda/aggregator/internal/handlers"

	// package drivers
	_ "github.com/lib/pq" // postgreSQL driver
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

	// open connection to PostgreSQL database
	db, err := sql.Open("postgres", *cfg.URL)
	// takes driver + db connection string (ptr to config.URL)

	// db check
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1) // clean exit
	}

	// create database instance
	dbQueries := database.New(db) // create db queries instance
	// dbQueries is a ptr to the Queries struct in the database package
	// provides methods to interact with the database instead of using raw SQL

	// create state instance and store config in
	state := &app.State{ // app
		Config: &cfg,
		DB:     dbQueries,
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

	// register the handler function for the login cmd
	cmds.Register("login", handlers.HandlerLogin)
	// Registers receivces commands
	// "login" = the command we register
	// HandlerLogin works on handlers, and registers "login" there

	// register the handler function for the register cmd
	cmds.Register("register", handlers.HandlerRegister)
	// Registers receivces commands
	// "register" = the command we register
	// HandlerRegister works on handlers, and registers "register" there

	// register the handler function for the reset cmd
	cmds.Register("reset", handlers.HandlerReset)
	// resets users table to prevent down/up migration for each test
	// "reset" = the command we register
	// HandlerReset works on handlers, and registers "reset" there

	// register the handler function for the users cmd
	cmds.Register("users", handlers.HandlerGetUsers)
	// prints all users in the database, and the current user
	// "users" = the command we register
	// HandlerReset works on handlers, and registers "users" there

	// register the handler function for the agg cmd
	cmds.Register("agg", handlers.HandlerAgg)
	// prints the aggregated RSS feed
	// "agg" = the command we register
	// HandlerAgg works on handlers, and registers "agg" there

	// register the handler function for the addfeed cmd
	cmds.Register("addfeed", handlers.MiddlewareLoggedIn(handlers.HandlerAddFeed))
	// adds a feed to the database (And follows it)
	// "addfeed" = the command we register
	// HandlerAgg works on handlers, and registers "addfeed" there

	// register the handler function for the feeds cmd
	cmds.Register("feeds", handlers.HandlerFeeds)
	// lists all feeds in db with created user name
	// "feeds" = the command we register
	// HandlerFeeds works on handlers, and registers "feeds" there

	// register the handler function for the follow cmd
	cmds.Register("follow", handlers.MiddlewareLoggedIn(handlers.HandlerFollow))
	// current user follows a feed
	// "follow" = the command we register
	// HandlerFollow works on handlers, and registers "follow" there

	// register the handler function for the following cmd
	cmds.Register("following", handlers.MiddlewareLoggedIn(handlers.HandlerFollowing))
	// lists all feeds the current user follows
	// "following" = the command we register
	// HandlerFollowing works on handlers, and registers "following" there

	// register the handler function for the unfollow cmd
	cmds.Register("unfollow", handlers.MiddlewareLoggedIn(handlers.HandlerUnfollow))
	// unfollows a feed for the logged in user
	// "unfollow" = the command we register
	// HandlerUnfollow works on handlers, and registers "unfollow" there

	// register the handler function for the unfollow cmd
	cmds.Register("browse", handlers.MiddlewareLoggedIn(handlers.HandlerBrowse))
	// browse lists all the posts for the logged in user
	// "browse" = the command we register
	// HandlerBrowse works on handlers, and registers "browse" there

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
