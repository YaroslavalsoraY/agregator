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
	"golang.org/x/net/html"
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
	url := "https://www.wagslane.dev/index.xml"

	urlData, err := fetchFeed(context.Background(), url)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("Error: %v", err)
	}

	fmt.Println(html.UnescapeString(urlData.Channel.Title))
	fmt.Println(html.UnescapeString(urlData.Channel.Description))
	
	for _, el := range urlData.Channel.Item {
		fmt.Println(html.UnescapeString(el.Title))
		fmt.Println(html.UnescapeString(el.Description))
	}
	
	return nil
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
