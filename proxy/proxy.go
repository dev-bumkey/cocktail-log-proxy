package proxy

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"Goproxy/auth"
	"Goproxy/config"
	"Goproxy/logger"
)

type ApiResponse struct {
	Status string       `json:"status"`
	Code   string       `json:"code"`
	Result []LogService `json:"result"`
}

type LogService struct {
	AccountSeq int `json:"accountSeq"`
	// LogServiceSeq int `json:"logServiceSeq"`
	ApiEndpointUrl string `json:"apiEndpointUrl"`
	Activated      string `json:"activated"`
}

var (
	logServiceData     ApiResponse
	logServiceDataLock sync.RWMutex
)

var startTime = time.Now()

// 주기적으로 데이터 업데이트
func UpdateLogServiceData(ctx context.Context) {
	updateData()

	// 주기 설정
	ticker := time.NewTicker(config.Data.UpdateCycle)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// api-server 주기적으로 호출
			updateData()
		}
	}
}

func updateData() {
	newProxyingUrl, err := fetchData(config.Data.CocktailApiUrl)
	if err != nil {
		log.Println("Error updating LogService data:", err)
		return
	}

	// Update logServiceData with the new data
	logServiceDataLock.Lock()
	logServiceData = newProxyingUrl
	logServiceDataLock.Unlock()
}

// api-server response url 가져오기
func fetchData(url string) (ApiResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return ApiResponse{}, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ApiResponse{}, err
	}

	var apiRes ApiResponse
	err = json.Unmarshal(responseBody, &apiRes)
	if err != nil {
		return ApiResponse{}, err
	}

	return apiRes, nil
}

// HTTP 요청을 처리하고 프록시 서버 역할
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Read from the updated logServiceData
	logServiceDataLock.RLock()
	defer logServiceDataLock.RUnlock()

	accountSeq, userID, userRole, err := auth.Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Find activated endpoint for the current account
	var activatedEndpointUrl string

	for _, logService := range logServiceData.Result {
		if logService.AccountSeq == accountSeq && logService.Activated == "Y" {
			activatedEndpointUrl = logService.ApiEndpointUrl
			break
		}
	}

	if len(activatedEndpointUrl) == 0 {
		http.Error(w, "No enabled URLs found for the given Account-Seq", http.StatusInternalServerError)
		return
	}

	// 요청 URL에 path add
	activatedEndpointUrl += r.URL.Path
	if r.URL.RawQuery != "" {
		activatedEndpointUrl += "?" + r.URL.RawQuery
	}

	// 활성화된 엔드포인트로 요청 프록시
	req, err := http.NewRequest(r.Method, activatedEndpointUrl, r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("account-seq", strconv.Itoa(accountSeq))
	req.Header.Set("user-id", userID)
	req.Header.Set("user-role", userRole)

	// Send request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error proxying request:", err)
		http.Error(w, "Error proxying request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Error copying response body", http.StatusInternalServerError)
		return
	}

	// Log response details
	logger.LogRequest(r, resp)
}
