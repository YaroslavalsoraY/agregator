package internal

import (
	"errors"
	"fmt"
	"main/internal/config"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

type State struct {
	ConfPtr *config.Config
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Not enough arguments")
	}

	err := s.ConfPtr.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	fmt.Println("Username has been set")
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
