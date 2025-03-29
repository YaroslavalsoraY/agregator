package config

import (
	"encoding/json"
	"errors"
	"os"
)

const Json_name = "/.gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func ReadConf() (Config, error){
	path, err := os.UserHomeDir()
	c := Config{}

	if err != nil {
		return c, errors.New("Problem with home path")
	}
	
	path += Json_name
	
	raw_json, err := os.ReadFile(path)
	if err != nil {
		return c, errors.New("Problem with reading")
	}
	
	err = json.Unmarshal(raw_json, &c)
	if err != nil {
		return c, errors.New("Problem with unmarshalling")
	}

	return c, nil
}

func (c *Config) SetUser(name string) error {
	path, err := os.UserHomeDir()
	if err != nil {
		return errors.New("Problem with home path")
	}
	path += Json_name

	c.Current_user_name = name

	data, err := json.Marshal(c)
	if err != nil {
		return errors.New("Problem with marshalling")
	}

	err = os.WriteFile(path, data, 0777)
	if err != nil {
		return errors.New("Problem with WriteFile")
	}

	return nil
}