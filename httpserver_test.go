package giniapi

import (
	// "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	// "io/ioutil"
	// "log"
	"net/http"
	"net/http/httptest"
	// "strconv"
	// "time"
)

var (
	testHTTPServer *httptest.Server
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/ping", handlerGetPing).Methods("GET")
	r.HandleFunc("/oauth/token", handlerPostToken).Methods("POST")
	r.HandleFunc("/test/http/basicAuth", handlerTestHttpBasicAuth).Methods("GET")
	r.HandleFunc("/test/http/oauth2", handlerTestHttpOauth2).Methods("GET")
	testHTTPServer = httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		// body, _ := ioutil.ReadAll(r.Body)
		// log.Printf("%s \"%s %s\" - %v => %v\n\n", r.RemoteAddr, r.Method, r.URL, r.Header, string(body))
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
}

func handlerGetPing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func handlerPostToken(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
                "access_token":"760822cb-2dec-4275-8da8-fa8f5680e8d4",
                "token_type":"bearer",
                "expires_in":300,
                "refresh_token":"46463dd6-cdbb-440d-88fc-b10a34f68b26"
             }`

	w.Write([]byte(body))
}

func handlerTestHttpBasicAuth(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/vnd.gini.v1+json" {
		writeHeaders(w, 500, "changes")
	} else {
		writeHeaders(w, 200, "changes")
	}
	body := "test completed"
	w.Write([]byte(body))
}

func handlerTestHttpOauth2(w http.ResponseWriter, r *http.Request) {
	body := "test completed"
	if r.Header.Get("Authorization") != "Bearer 760822cb-2dec-4275-8da8-fa8f5680e8d4" {
		writeHeaders(w, 401, "invalid token")
		body = "Invalid Authorization header"
	} else {
		writeHeaders(w, 200, "changes")
	}
	w.Write([]byte(body))
}

func writeHeaders(w http.ResponseWriter, code int, jobName string) {
	h := w.Header()
	h.Add("Content-Type", "application/json")
	if jobName != "" {
		h.Add("Job-Name", jobName)
	}
	w.WriteHeader(code)
}
