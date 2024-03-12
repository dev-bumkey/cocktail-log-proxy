package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
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

var startTime = time.Now()

// ProxyHandler : HTTP 요청을 처리하고 프록시 서버역할
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// GET 요청을 보내서 LogService JSON 데이터 가져오기
	resp, err := http.Get("http://localhost:8080/internal/log/service")
	if err != nil {
		http.Error(w, "Error sending GET request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	accountSeq := r.Header.Get("account-seq")
	userID := r.Header.Get("user-id")
	userRole := r.Header.Get("user-role")

	if userID == "" || userRole == "" || accountSeq == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	seq, err := strconv.Atoi(accountSeq)
	if err != nil {
		http.Error(w, "Invalid account-Seq header", http.StatusBadRequest)
		return
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Response 값 변환
	jsonResponse := LogServiceResponse{}

	err = json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		fmt.Errorf("fail to read response: %s", err.Error())
		http.Error(w, "fail to read response", http.StatusInternalServerError)
		return
	}

	activatedEndpoint := ""
	for _, logService := range jsonResponse.Result {
		// LogService에서 accountSeq가 header.accountSeq이고 activated가 "Y"인 경우의 apiEndpointUrl을 찾기
		if logService.AccountSeq == seq && logService.Activated == "Y" {
			activatedEndpoint = logService.ApiEndpointUrl
			break
		}
	}

	// 프록싱된 결과를 클라이언트에게 반환
	if activatedEndpoint != "" {
		http.Redirect(w, r, activatedEndpoint, http.StatusFound)
	} else {
		// 현재 계정에 대한 활성화된 로그 서비스를 찾을 수 없을때
		http.Error(w, "No active log services were found for your current account.", http.StatusNotFound)
		return
	}

	// 프록싱된 url check
	resp, err = http.Get(activatedEndpoint)
	if err != nil {
		// 프록시된 요청이 실패한 경우
		log.Println("Error proxying request:", err.Error())
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

	// Response log
	elapsed := time.Since(startTime)

	// r.RemoteAddr: 요청이 온 클라이언트의 IP 주소와 포트
	// activatedEndpoint: 프록시된 요청의 대상 엔드포인트
	// r.Method r.URL.Path r.Proto: 요청의 HTTP 메소드, URL 경로 및 프로토콜
	// resp.StatusCode: 응답의 HTTP 상태 코드
	// resp.ContentLength: 응답 본문의 길이
	// r.Header.Get("X-Forwarded-For"): 요청 헤더의 "X-Forwarded-For"
	log.Printf(
		"%s %s \"%s\" %d %d \"%s\" %.3fms\n",
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
