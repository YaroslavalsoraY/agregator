package main

import (
	"log"
	"main/internal"
	"main/internal/config"
	"os"
)

func main() {
	conf, err := config.ReadConf()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		os.Exit(1)
	}
	currentState := internal.State{ConfPtr: &conf}
	cmdList := internal.Commands{Handlers: make(map[string]func(*internal.State, internal.Command) error)}
	cmdList.Register(os.Args[0], internal.HandlerLogin)
	if len(os.Args) < 2 {
		log.Fatalf("not enough arguments")
		os.Exit(1)
	}
	cmdName := os.Args[0]
	cmdArgs := os.Args[2:]
	cmdSign := internal.Command{Name: cmdName, Args: cmdArgs}
	err = cmdList.Run(&currentState, cmdSign)
	if err != nil {
		log.Fatalf("error with running: %v", err)
		os.Exit(1)
	}
}