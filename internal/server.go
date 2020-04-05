package internal

import (
	"encoding/json"
	"fmt"
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
	NewRemindHandler(db, r.PathPrefix("/remind").Subrouter())

	// middleware
	r.Use(s.getLoggingMiddleware())

	return r
}

func sendResp(w http.ResponseWriter, statusCode int, response Response) {
	w.WriteHeader(statusCode)
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

func sendSimpleResponse(w http.ResponseWriter, message string) {
	sendResp(w, http.StatusOK, Response{
		Ok:      true,
		Message: message,
		Data:    nil,
	})
}

func sendSimpleErrResponse(w http.ResponseWriter, statusCode int, message string) {
	sendResp(w, statusCode, Response{
		Ok:      false,
		Message: message,
		Data:    nil,
	})
}

func sendSimpleBadRequestResponse(w http.ResponseWriter, message string) {
	sendResp(w, http.StatusBadRequest, Response{
		Ok:      false,
		Message: message,
		Data:    nil,
	})
}

func (s *Server) getLoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAgent := r.Header.Get("User-Agent")
			log.Tracef(" ====> request [%s] path: [%s] [UA: %s]", r.Method, r.URL.Path, userAgent)
			next.ServeHTTP(w, r)
		})
	}
}
