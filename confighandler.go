package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Callbackuri  string
	Clientid     string
	Clientsecret string
	Databaseurl  string
}

var config Config

func ConfigInit() {
	//Check if file exists anton_config.json
	if _, err := os.Stat("./chat-overlay-api.json"); os.IsNotExist(err) {
		log.Println("chat-overlay-api.json did not exist and have been created. Please fill in the fields")
		file, _ := json.MarshalIndent(config, "", " ")

		_ = ioutil.WriteFile("./chat-overlay-api.json", file, 0644)
		os.Exit(1)
	}

	//Load config.
	file, err := os.Open("./chat-overlay-api.json")
	if err != nil {
		log.Println(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Println(err)
	}

	//log.Println(config)
}
