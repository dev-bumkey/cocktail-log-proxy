package main

import (
	"fmt"
	"log"
	"net/http"

	"Goproxy/proxy"
)

func main() {
	// ProxyHandler 함수를 요청 핸들러로 등록
	http.HandleFunc("/", proxy.ProxyHandler)

	// 9999 포트에서 서버 시작
	fmt.Println("Starting proxy server on :9999")
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		log.Fatal("Error starting proxy server:", err)
	}
}
