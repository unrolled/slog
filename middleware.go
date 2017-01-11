package slog

import "net/http"

var (
	// RequestHeaderKey is the key used when adding the header token.
	RequestHeaderKey = "___slog_request_token___"

	// RequestFieldKey is the key used in the Field output.
	RequestFieldKey = "reqID"

	// ResponseHeaderKey will send the token on the response as well (if set).
	ResponseHeaderKey = ""
)

// Requestify adds a unique key to the request (header) and uses it for logging.
func Requestify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := RequestToken.Generate()
		r.Header.Add(RequestHeaderKey, token)

		if RequestFieldKey != "" {
			w.Header().Add(ResponseHeaderKey, token)
		}
		next.ServeHTTP(w, r)
	})
}
