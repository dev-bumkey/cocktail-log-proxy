package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var customTransport = http.DefaultTransport

var configFileMutex sync.Mutex

var startTime = time.Now()

type AccountInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func GetEnabledURLs(filePath string, seq int) ([]string, error) {
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

func HandleRequest(w http.ResponseWriter, r *http.Request, configFile string) {
	accountSeq := r.Header.Get("Account-Seq")
	userID := r.Header.Get("user-id")
	userRole := r.Header.Get("user-role")

	if userID == "" || userRole == "" || accountSeq == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	seq, err := strconv.Atoi(accountSeq)
	if err != nil {
		http.Error(w, "Invalid Account-Seq header", http.StatusBadRequest)
		return
	}

	configFileMutex.Lock()
	defer configFileMutex.Unlock()

	// config파일 못불러옴
	enabledURLs, err := GetEnabledURLs(configFile, seq)
	if err != nil {
		fmt.Println(enabledURLs, " enabledURLs ")
		log.Println("Error getting enabled URLs:", err.Error())
		http.Error(w, "Error getting enabled URLs", http.StatusInternalServerError)
		return
	}

	var targetURL string
	if len(enabledURLs) > 0 {
		targetURL = enabledURLs[0]
		// Add Path from the original request URL
		targetURL += r.URL.Path

		// fmt.Println("targetURL:", targetURL)
		// fmt.Println("path:", r.URL.Path)
		// fmt.Println("r.URL.RawQuery:", r.URL.RawQuery)
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}
	} else {
		http.Error(w, "No enabled URLs found for the given Account-Seq", http.StatusInternalServerError)
		return
	}

	// server 켜있는지 확인
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
		log.Println("Error sending proxy request:", err.Error())
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)

	log.Printf(
		"%s %s \"%s\" %d %d \"%s\" %.3fms\n",
		r.RemoteAddr,
		targetURL,
		r.Method+" "+r.URL.Path+" "+r.Proto,
		resp.StatusCode,
		resp.ContentLength,
		r.Header.Get("X-Forwarded-For"),
		float64(elapsed.Microseconds())/1000000.0,
	)

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}
