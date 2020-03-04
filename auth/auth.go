// Package auth defines server functions for authenticating incoming requests.
package auth

import (
	"crypto/hmac"
	"fmt"
	"net/http"

	"github.com/izzymg/rotcommon"
)

/*
AuthHeader is the key of the header which contains the
authentication hash. */
const AuthHeader = "Rot-Authorization"

/*
TokenHeader is the key of the header which contains the
payload that was hashed for auth. */
const TokenHeader = "Rot-Token"

// MissingAuthMessage is a string sent when no token was provided to a request.
const MissingAuthMessage = "Missing auth"

// MissingTokenMessage is a string sent when no token was provided to a request.
const MissingTokenMessage = "Missing token"

// Return a handler that writes an HTTP 403 with a plain text message.
func forbid(message string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("content-type", "text/plain")
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(rw, "%s", message)
	})
}

// Middleware authenticates the request with a hash before calling next.
func Middleware(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get(AuthHeader)
		tokenHeader := req.Header.Get(TokenHeader)

		if len(authHeader) < 1 {
			forbid(MissingAuthMessage).ServeHTTP(rw, req)
			return
		}

		if len(tokenHeader) < 1 {
			forbid(MissingTokenMessage).ServeHTTP(rw, req)
			return
		}

		digest, err := rotcommon.HashData(tokenHeader, secret)
		if err != nil {
			forbid(MissingAuthMessage).ServeHTTP(rw, req)
			return
		}

		ok := hmac.Equal([]byte(digest), []byte(authHeader))
		if !ok {
			forbid(MissingAuthMessage).ServeHTTP(rw, req)
			return
		}

		next.ServeHTTP(rw, req)
	})
}
