package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DB_PASS string
	DB_USER string
	DB_NAME string
}

func LoadConfig(configFilename string) Config {
	file, _ := os.Open(configFilename)
	decoder := json.NewDecoder(file)
	config := Config{}

	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
	}

	return config
}
