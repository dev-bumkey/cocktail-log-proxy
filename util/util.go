package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AccountInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func GetEnabledURLs(filePath string, seq int) ([]string, error) {
	// JSON 파일 읽기
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var accounts map[string][]AccountInfo
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	var enabledURLs []string
	seqStr := strconv.Itoa(seq)
	for accountId, accountInfo := range accounts {
		if strings.HasSuffix(accountId, seqStr) {
			for _, info := range accountInfo {
				if info.Enabled {
					enabledURLs = append(enabledURLs, info.URL)
				}
			}
		}
	}
	return enabledURLs, nil
}
