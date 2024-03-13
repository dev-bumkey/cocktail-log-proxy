package logger

import (
	"log"
	"net/http"
	"time"
)

var startTime = time.Now()

func LogRequest(r *http.Request, resp *http.Response) {

	elapsed := time.Since(startTime)

	// Log response details
	// r.RemoteAddr: 요청이 온 클라이언트의 IP 주소와 포트
	// activatedEndpointUrl: 프록시된 요청의 대상 엔드포인트
	// r.Method r.URL.Path r.Proto: 요청의 HTTP 메소드, URL 경로 및 프로토콜
	// resp.StatusCode: 응답의 HTTP 상태 코드
	// resp.ContentLength: 응답 본문의 길이
	// r.Header.Get("X-Forwarded-For"): 요청 헤더의 "X-Forwarded-For"

	log.Printf(
		"%s %s \"%s\" %d %d \"%s\" %.6f",
		r.RemoteAddr,
		resp.Request.URL.String(),
		resp.Request.Method+" "+resp.Request.URL.Path+" "+resp.Request.Proto,
		resp.StatusCode,
		resp.ContentLength,
		r.Header.Get("X-Forwarded-For"),
		float64(elapsed.Microseconds())/1000000.0,
	)
}
