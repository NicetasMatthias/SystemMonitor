package server

import (
	"fmt"
	"strconv"
)

type Config struct {
	Port string `json:"port"`
}

func DefaultConfig() Config {
	return Config{
		Port: "55432",
	}
}

func (c *Config) Validate() error {

	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("failed to convert server port to integer: %v", err)
	}

	if port <= 0 {
		return fmt.Errorf("server port must be a string containing a positive integer, got = %s", c.Port)
	}

	return nil
}
