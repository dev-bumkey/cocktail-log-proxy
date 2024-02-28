package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"Goproxy/util"
)

var customTransport = http.DefaultTransport

func init() {
	// Here, you can customize the transport, e.g., set timeouts or enable/disable keep-alive
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 헤더에서 Account-Seq 추출
	accountSeq := r.Header.Get("Account-Seq")

	seq, err := strconv.Atoi(accountSeq)
	if err != nil {
		http.Error(w, "Invalid Account-Seq header", http.StatusBadRequest)
		return
	}

	// filepath := "/var/conf/config.json"
	filepath := "config.json"

	// Get enabled URLs based on Account-Seq
	enabledURLs, err := util.GetEnabledURLs(filepath, seq)
	if err != nil {
		http.Error(w, "Error getting enabled URLs", http.StatusInternalServerError)
		return
	}
	fmt.Println("enabledURLs", enabledURLs)

	var targetURL string
	if len(enabledURLs) > 0 {
		// 첫번째 객체만 반환 => 추후 바꿔야함
		targetURL = enabledURLs[0]

		// Add Path from the original request URL
		// targetURL += r.URL.Path

		//
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}
		// if r.URL.Path != "" {
		// 	targetURL += r.URL.Path
		// }
	} else {
		http.Error(w, "No enabled URLs found for the given Account-Seq", http.StatusInternalServerError)
		return
	}

	// Logging request message
	fmt.Println("Request targetURL:", targetURL)
	fmt.Println("Request URL:", r.URL)
	fmt.Println("Request URL.path:", r.URL.Path)

	fmt.Println("----------------------")

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Send the proxy request using the custom transport
	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
		fmt.Println("Error sending proxy request:", err.Error())
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Logging response status code
	fmt.Println("Response Status Code:", resp.StatusCode)
	fmt.Println("Response Header:", resp.Header)
	fmt.Println("Response Body:", resp.Body)

	fmt.Println("----------------------")
	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	io.Copy(w, resp.Body)
}
func main() {
	// Create a new HTTP server with the handleRequest function as the handler
	server := http.Server{
		Addr:    ":9999",
		Handler: http.HandlerFunc(handleRequest),
	}

	// Start the server and log any errors
	log.Println("Starting proxy server on :9999")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}
