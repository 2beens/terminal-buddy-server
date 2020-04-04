package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
}

func NewServer() *Server {
	server := &Server{}
	// TODO:
	return server
}

func (s *Server) Serve(port string) {
	memDb := NewMemDb()
	router := s.routerSetup(memDb)

	ipAndPort := fmt.Sprintf("%s:%s", "localhost", port)
	httpServer := &http.Server{
		Handler:      router,
		Addr:         ipAndPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof(" > server listening on: [%s]", ipAndPort)
	log.Fatal(httpServer.ListenAndServe())
}

func (s *Server) routerSetup(db BuddyDb) *mux.Router {
	log.Trace("setting routes")
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sendSimpleResponse(w, "WIP")
	})

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		sendSimpleResponse(w, "i'm fine <3")
	})

	// handle register
	NewUserHandler(db, r.PathPrefix("/user").Subrouter())

	// handle remind

	// middleware
	r.Use(s.getLoggingMiddleware())

	return r
}

func sendResp(w io.Writer, response Response) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Warnf("#120412 failed to send response: %s", err)
		return
	}
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Warnf("#120413 failed to send response: %s", err)
		return
	}
}

func sendSimpleResponse(w io.Writer, message string) {
	sendResp(w, Response{
		Ok:      true,
		Message: message,
		Data:    nil,
	})
}

func sendSimpleErrResponse(w io.Writer, message string) {
	sendResp(w, Response{
		Ok:      false,
		Message: message,
		Data:    nil,
	})
}

func (s *Server) getLoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAgent := r.Header.Get("User-Agent")
			sessionID := r.Header.Get("X-Ispend-SessionID")
			log.Tracef(" ====> request [%s] path: [%s] [sessionID: %s] [UA: %s]", r.Method, r.URL.Path, sessionID, userAgent)
			next.ServeHTTP(w, r)
		})
	}
}
