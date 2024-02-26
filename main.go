package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type AccountInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func main() {
	// JSON 파일 읽기
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var accounts map[string][]AccountInfo
	if err := json.Unmarshal(data, &accounts); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 각 계정에 대한 정보 확인
	for accountId, accountInfo := range accounts {
		fmt.Printf("Account ID: %s\n", accountId)
		for _, info := range accountInfo {
			if info.Enabled {
				fmt.Printf("Name: %s\nURL: %s\n\n", info.Name, info.URL)
			}
		}
	}
}
