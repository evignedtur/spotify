package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var mu sync.Mutex

func writeTokensToFile() {
	createFileIfNotExist()

	mu.Lock()
	defer mu.Unlock()

	file, _ := json.MarshalIndent(Tokens, "", " ")
	_ = ioutil.WriteFile("tokens.json", file, 0644)

}

func readTokensFromFile() {
	createFileIfNotExist()
	//Load Tokens

	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open("./tokens.json")
	if err != nil {
		log.Println(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Tokens)
	if err != nil {
		log.Println(err)
	}

	log.Println(Tokens)
}

func createFileIfNotExist() {
	if _, err := os.Stat("./tokens.json"); os.IsNotExist(err) {
		log.Println("tokens.json did not exist and have been created")

		_ = ioutil.WriteFile("./tokens.json", nil, 0644)
	}
}
