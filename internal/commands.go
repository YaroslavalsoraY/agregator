package internal

import (
	"context"
	"errors"
	"fmt"
	"main/internal/config"
	"main/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

type State struct {
	DB		*database.Queries
	ConfPtr *config.Config
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Not enough arguments")
	}

	_, err := s.DB.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		os.Exit(1)
		return errors.New("You can't login to an account that doesn't exist!")
	}

	err = s.ConfPtr.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	fmt.Println("Username has been set")
	return nil
}

func HandlerRegister(s *State, cmd Command) error{
	if len(cmd.Args) == 0{
		return errors.New("Not enoughn arguments")
	}
	_, err := s.DB.GetUser(context.Background(), cmd.Args[0])
	if err == nil {
		os.Exit(1)
		return errors.New("This name already exists")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(), 
		CreatedAt: time.Now(), 
		UpdatedAt: time.Now(), 
		Name:      cmd.Args[0],
	}
	user, err := s.DB.CreateUser(context.Background(), params)
	err = s.ConfPtr.SetUser(user.Name)
	if err != nil {
		return errors.New("Problem with writing in db")
	}
	fmt.Println("User is created. Welcome!")
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.Reset(context.Background())
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("error: %v", err)
	}
	fmt.Println("Database reset successfully!")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	namesList, err := s.DB.GetAllUsers(context.Background())
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	for _, el := range namesList {
		if el == s.ConfPtr.Current_user_name {
			fmt.Printf("* %s (current)\n", el)
		} else {
			fmt.Printf("* %s\n", el)
		}
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		os.Exit(1)
		return errors.New("Not enough arguments")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			os.Exit(1)
			return fmt.Errorf("Error: %v", err)
		}
	}
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		os.Exit(1)
		return errors.New("Not enough arguments")
	}

	userID, err := s.DB.GetUserID(context.Background(), s.ConfPtr.Current_user_name)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	arg := database.AddFeedParams {
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: userID,
	}

	newFeed, err := s.DB.AddFeed(context.Background(), arg)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}
	
	arg2 := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: userID,
		FeedID: newFeed.ID,
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), arg2)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	printFeed(newFeed)
	
	return nil
}

func HandlerFeeds(s *State, cmd Command, user database.User) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	for _, el := range feeds {
		fmt.Printf("Name: %s\nUrl: %s\nUsername: %s\n\n", el.Name, el.Url, el.Name_2)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		os.Exit(1)
		return errors.New("Not enough arguments")
	}

	userID, err := s.DB.GetUserID(context.Background(), s.ConfPtr.Current_user_name)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	feedID, err := s.DB.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	arg := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: userID,
		FeedID: feedID.ID,
	}
	_, err = s.DB.CreateFeedFollow(context.Background(), arg)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	fmt.Printf("%s followed on %s\n", feedID.Name, s.ConfPtr.Current_user_name)
	return nil
}

func HandlerFolowing(s *State, cmd Command) error {
	userID, err := s.DB.GetUserID(context.Background(), s.ConfPtr.Current_user_name)
	var feedNames []string

	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	feedIDs, err := s.DB.GetAllUsersFeedsIDs(context.Background(), userID)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}
	
	for _, feedID := range feedIDs {
		feedName, _ := s.DB.GetFeedByID(context.Background(), feedID)
		feedNames = append(feedNames, feedName)
	} 

	fmt.Println("All your feeds:")
	for _, el := range feedNames {
		fmt.Printf("%s\n", el)
	}
	
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		os.Exit(1)
		return errors.New("Not enough arguments")
	}

	feedID, err := s.DB.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	arg := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID.ID,
	}
	err = s.DB.DeleteFeedFollow(context.Background(), arg)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}
	
	return nil
}

func scrapeFeeds(s *State) (error) {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	err = s.DB.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	updatedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	fmt.Println(updatedFeed.Channel.Title)
	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
	fmt.Printf("* lstFetched:        %v\n", feed.LastFetchedAt)
}

func (c *Commands) Register(name string, f func(s *State, cmd Command) error) {
	c.Handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	err := c.Handlers[cmd.Name](s, cmd)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	return nil
}
