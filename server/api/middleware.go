package api

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
)

const (
	authHeader       string = "Authorization"
	authHeaderPrefix string = "Basic "
)

type middleware func(next http.Handler) http.Handler

func applyMiddleware(middlewares ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

type responseWithStatus struct {
	StatusCode int
	http.ResponseWriter
}

func (r *responseWithStatus) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func logging() middleware {
	return func(next http.Handler) http.Handler {
		now := time.Now()
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rsp := &responseWithStatus{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}
			next.ServeHTTP(rsp, r)
			log.Printf("[%s] %s - %d %s", r.Method, r.URL, rsp.StatusCode, time.Since(now))
		})
	}
}

func auth(serverConfig config.ServerConfig) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerValue := r.Header.Get(authHeader)
			if len(headerValue) == 0 || !strings.HasPrefix(headerValue, authHeaderPrefix) {
				http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
				return
			}

			authData := strings.TrimPrefix(headerValue, authHeaderPrefix)
			data, err := base64.StdEncoding.DecodeString(authData)
			if err != nil {
				log.Println(err)
				http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
				return
			}
			credentials := strings.Split(string(data), ":")
			if len(credentials) != 2 {
				http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
				return
			}

			if credentials[0] == serverConfig.Username && credentials[1] == serverConfig.Password {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
			}
		})
	}
}
