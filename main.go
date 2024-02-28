package main

import (
	"log"
	"net/http"

	"Goproxy/proxy"
)

func main() {
	// configFile := "config.json"
	configFile := "/var/conf/config.json"

	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			proxy.HandleRequest(w, r, configFile)
		},
	)

	log.Println("Starting proxy server on :9999")
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}
