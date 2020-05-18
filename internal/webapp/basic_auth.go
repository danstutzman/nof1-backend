package webapp

import (
	"net/http"
)

func basicAuth(h http.Handler, expectedUsername,
	expectedPassword string) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()

		if username == expectedUsername && password == expectedPassword {
			h.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin realm"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}
