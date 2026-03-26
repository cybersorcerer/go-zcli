package notifications

import (
	"net/http"
	"regexp"
)

var (
	NotifyRe = regexp.MustCompile(`^/notifications*$`)
)

type homeHandler struct{}
type notificationHandler struct{}

func (h *homeHandler) jobNotification(w http.ResponseWriter, r *http.Request)         {}
func (h *notificationHandler) jobNotification(w http.ResponseWriter, r *http.Request) {}

func notivicationServer() {
	// Create a new request multiplexer
	// Take incoming requests and dispatch them to the matching handlers
	mux := http.NewServeMux()
	// Register the routes and handlers
	mux.Handle("/", &homeHandler{})
	mux.Handle("/notifications", &notificationHandler{})
	// Run the server
	http.ListenAndServe("192.168.1.153:8080", mux)
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to zcli Submit Notification Server!"))
	w.Write([]byte("Endpoint Hit: homePage"))
}

func (h *notificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && NotifyRe.MatchString(r.URL.Path):
		h.jobNotification(w, r)
		return
	default:
		return
	}
}
