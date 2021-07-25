package auth

import (
	"base-service/app/baseapp/interface/internal/pkg/token"
	"encoding/json"
	"net/http"
)

type unAuthResult struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func PrimaryFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("AuthToken")
		if uid, ok := token.Verify(t); ok {
			r.Header.Set("uid", uid)
		}
		next.ServeHTTP(w, r)
	})
}

func RequireFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("AuthToken")
		if uid, ok := token.Verify(t); ok {
			r.Header.Set("uid", uid)
			next.ServeHTTP(w, r)
		} else {
			// 否则将直接返回
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			data, _ := json.Marshal(unAuthResult{
				Code:    401,
				Message: "Unauthorized request",
			})
			_, _ = w.Write(data)
		}
	})
}