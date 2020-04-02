package internal

import (
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
		w.Write([]byte("nice"))
	})

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i am fine :)"))
	})

	// handle register
	userRouter := r.PathPrefix("/user").Subrouter()
	NewUserHandler(db, userRouter)

	// handle remind

	return r
}
