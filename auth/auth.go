package auth

import (
	"net/http"
	"strconv"
)

func Authenticate(r *http.Request) (int, string, string, error) {
	accountSeq := r.Header.Get("account-seq")
	userID := r.Header.Get("user-id")
	userRole := r.Header.Get("user-role")

	// Check if required header values are present
	if userID == "" || userRole == "" || accountSeq == "" {
		return 0, "", "", http.ErrNoCookie
	}

	seq, err := strconv.Atoi(accountSeq)
	if err != nil {
		return 0, "", "", err
	}

	return seq, userID, userRole, nil
}
