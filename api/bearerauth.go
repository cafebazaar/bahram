package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
)

var tokenEntropy = 32

// AuthBearerMiddleware provides a Token Auth implementation. On success, the wrapped middleware
// is called, and the userID is made available as request.Env["REMOTE_USER"].(string)
type AuthBearerMiddleware struct {
	// Realm name to display to the user. Required.
	Realm string

	// Callback function that should perform the authentication of the user based on token.
	// Must return userID as string on success, empty string on failure. Required.
	// The returned userID is normally the primary key for your user record.
	Authenticator func(token string) string

	// Callback function that should perform the authorization of the authenticated user.
	// Must return true on success, false on failure. Optional, defaults to success.
	// Called only after an authentication success.
	Authorizer func(request *rest.Request) bool
}

// MiddlewareFunc makes AuthBearerMiddleware implement the Middleware interface.
func (mw *AuthBearerMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	if mw.Realm == "" {
		log.Fatal("Realm is required")
	}

	if mw.Authenticator == nil {
		log.Fatal("Authenticator is required")
	}

	if mw.Authorizer == nil {
		mw.Authorizer = func(request *rest.Request) bool {
			return true
		}
	}

	return func(writer rest.ResponseWriter, request *rest.Request) {
		authHeader := request.Header.Get("Authorization")

		if request.Method == "OPTIONS" {
			handler(writer, request)
			return
		}

		// Authorization header was not provided
		if authHeader == "" {
			mw.unauthorized(writer)
			return
		}

		token, err := decodeAuthHeader(authHeader)
		// Authorization header was *malformed* such that we couldn't extract a token
		if err != nil {
			mw.unauthorized(writer)
			return
		}

		userID := mw.Authenticator(token)
		// The token didn't map to a user, it's most likely either invalid or expired
		if userID == "" {
			mw.unauthorized(writer)
			return
		}

		// The user's token was valid, but they're not authorized for the current request
		if !mw.Authorizer(request) {
			mw.unauthorized(writer)
			return
		}

		request.Env["REMOTE_USER"] = userID
		handler(writer, request)
	}
}

func (mw *AuthBearerMiddleware) unauthorized(writer rest.ResponseWriter) {
	writer.Header().Set("WWW-Authenticate", "Token realm="+mw.Realm)
	rest.Error(writer, "Not Authorized", http.StatusUnauthorized)
}

// Extract the token from an Authorization header
func decodeAuthHeader(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return "", errors.New("Invalid Authorization header")
	}
	return string(parts[1]), nil
}
