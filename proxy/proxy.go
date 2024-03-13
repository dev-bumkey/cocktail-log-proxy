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

	"Goproxy/config"
)

type LogServiceResponse struct {
	Status string       `json:"status"`
	Code   string       `json:"code"`
	Result []LogService `json:"result"`
}

type LogService struct {
	AccountSeq     int    `json:"accountSeq"`
	LogServiceSeq  int    `json:"logServiceSeq"`
	LogServiceId   string `json:"logServiceId"`
	LogServiceType string `json:"logServiceType"`
	LogServiceName string `json:"logServiceName"`
	ApiEndpointUrl string `json:"apiEndpointUrl"`
	EndpointType   string `json:"endpointType"`
	HostIp         string `json:"hostIp"`
	ClusterSeq     int    `json:"clusterSeq"`
	ClusterId      string `json:"clusterId"`
	Namespace      string `json:"namespace"`
	Activated      string `json:"activated"`
}

var (
	logServiceData     LogServiceResponse
	logServiceDataLock sync.RWMutex
)

var startTime = time.Now()

// 주기적으로 데이터 업데이트
func UpdateLogServiceData(ctx context.Context) {
	// 주기 설정
	ticker := time.NewTicker(config.Data.UpdateCycle)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// api-server 주기적으로 호출
			newData, err := fetchData(config.Data.CocktailApiUrl)
			// fmt.Println("호출됨")
			if err != nil {
				log.Println("Error updating LogService data:", err)
				continue
			}

			// Update logServiceData with the new data
			logServiceDataLock.Lock()
			logServiceData = newData
			logServiceDataLock.Unlock()
		}
	}
}

// api-server response url 가져오기
func fetchData(url string) (LogServiceResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return LogServiceResponse{}, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return LogServiceResponse{}, err
	}

	var jsonResponse LogServiceResponse
	err = json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		return LogServiceResponse{}, err
	}

	return jsonResponse, nil
}

// ProxyHandler HTTP 요청을 처리하고 프록시 서버 역할
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Read from the updated logServiceData
	logServiceDataLock.RLock()
	defer logServiceDataLock.RUnlock()

	accountSeq := r.Header.Get("account-seq")
	userID := r.Header.Get("user-id")
	userRole := r.Header.Get("user-role")

	// Check if required header values are present
	if userID == "" || userRole == "" || accountSeq == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	seq, err := strconv.Atoi(accountSeq)
	if err != nil {
		http.Error(w, "Invalid account-Seq header", http.StatusBadRequest)
		return
	}

	// Find activated endpoint for the current account
	var activatedEndpoint string
	for _, logService := range logServiceData.Result {
		if logService.AccountSeq == seq && logService.Activated == "Y" {
			activatedEndpoint = logService.ApiEndpointUrl
			break
		}
	}

	// 현재 계정에 대한 활성화된 로그 서비스를 찾을 수 없을때
	if activatedEndpoint == "" {
		http.Error(w, "No active log services were found for your current account.", http.StatusNotFound)
		return
	}

	// Proxy the request to the activated endpoint
	resp, err := http.Get(activatedEndpoint)
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

	// Set response status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Error copying response body", http.StatusInternalServerError)
		return
	}

	// Log response details
	// r.RemoteAddr: 요청이 온 클라이언트의 IP 주소와 포트
	// activatedEndpoint: 프록시된 요청의 대상 엔드포인트
	// r.Method r.URL.Path r.Proto: 요청의 HTTP 메소드, URL 경로 및 프로토콜
	// resp.StatusCode: 응답의 HTTP 상태 코드
	// resp.ContentLength: 응답 본문의 길이
	// r.Header.Get("X-Forwarded-For"): 요청 헤더의 "X-Forwarded-For"
	elapsed := time.Since(startTime)

	log.Printf(
		"%s %s \"%s\" %d %d \"%s\"",
		r.RemoteAddr,
		activatedEndpoint,
		r.Method+" "+r.URL.Path+" "+r.Proto,
		resp.StatusCode,
		resp.ContentLength,
		r.Header.Get("X-Forwarded-For"),
		float64(elapsed.Microseconds())/1000000.0,
	)

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
}
