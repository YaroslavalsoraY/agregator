package main

import (
	"context"
	"database/sql"
	"log"
	"main/internal"
	"main/internal/config"
	"main/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	conf, err := config.ReadConf()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		os.Exit(1)
	}
	
	db, err := sql.Open("postgres", conf.DbURL)
	dbQueries := *database.New(db)
	if err != nil {
		log.Fatalf("error with database: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	
	currentState := internal.State{
		ConfPtr: &conf, 
		DB: &dbQueries,
	}
	
	cmdList := internal.Commands{
		Handlers: make(map[string]func(*internal.State, internal.Command) error),
	}
	
	cmdList.Register("login", internal.HandlerLogin)
	cmdList.Register("register", internal.HandlerRegister)
	cmdList.Register("reset", internal.HandlerReset)
	cmdList.Register("users", internal.HandlerUsers)
	cmdList.Register("agg", internal.HandlerAgg)
	cmdList.Register("addfeed", middlewareLoggedIn(internal.HandlerAddFeed))
	cmdList.Register("feeds", middlewareLoggedIn(internal.HandlerFeeds))
	cmdList.Register("follow", middlewareLoggedIn(internal.HandlerFollow))
	cmdList.Register("following", internal.HandlerFolowing)
	cmdList.Register("unfollow", middlewareLoggedIn(internal.HandlerUnfollow))

	if len(os.Args) < 2 {
		log.Fatalf("not enough arguments")
		os.Exit(1)
	}
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	cmdSign := internal.Command{Name: cmdName, Args: cmdArgs}
	
	err = cmdList.Run(&currentState, cmdSign)
	if err != nil {
		log.Fatalf("error with running: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func middlewareLoggedIn(handler func(s *internal.State, cmd internal.Command, user database.User) error) func(*internal.State, internal.Command) error {
	return func(s *internal.State, cmd internal.Command) error {
		user, err := s.DB.GetUser(context.Background(), s.ConfPtr.Current_user_name)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}