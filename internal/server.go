package internal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	db *pg.DB
}

func NewServer(recreateDb bool) *Server {
	server := &Server{}

	server.db = pg.Connect(&pg.Options{
		ApplicationName: "terminal-buddy",
		User:            "termbuddy",
		Database:        "termbuddydb",
	})

	if !server.DbOk() {
		panic("DB connection not happy ...")
	}

	err := server.createSchema(recreateDb)
	if err != nil {
		panic(err)
	}

	return server
}

func (s *Server) createSchema(recreateDb bool) error {
	if recreateDb {
		for _, model := range []interface{}{(*User)(nil), (*Reminder)(nil)} {
			err := s.db.DropTable(model, &orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, model := range []interface{}{(*User)(nil), (*Reminder)(nil)} {
		err := s.db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	admin := &User{
		Username:     "serj",
		PasswordHash: fmt.Sprintf("%x", md5.Sum([]byte("serj"))),
		Reminders:    nil,
	}

	err := s.db.Insert(admin)
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *Server) DbOk() bool {
	_, err := s.db.Exec("SELECT 1")
	if err != nil {
		return false
	}
	return true
}

func (s *Server) Serve(port int) {
	memDb := NewMemDb()
	router := s.routerSetup(memDb)

	ipAndPort := fmt.Sprintf("%s:%d", "localhost", port)
	httpServer := &http.Server{
		Handler:      router,
		Addr:         ipAndPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	chOsInterrupt := make(chan os.Signal, 1)
	signal.Notify(chOsInterrupt, os.Interrupt)

	go func() {
		log.Infof(" > server listening on: [%s]", ipAndPort)
		log.Fatal(httpServer.ListenAndServe())
	}()

	select {
	case <-chOsInterrupt:
		log.Warn("os interrupt received!")
	}
	s.shutdown()
}

func (s *Server) shutdown() {
	log.Debugf("shutting down DB ...")
	if err := s.db.Close(); err != nil {
		log.Errorf("failed to close DB connection: %s", err.Error())
	} else {
		log.Debugf("DB shut down")
	}
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
		Ok:            true,
		Message:       message,
		DataJsonBytes: nil,
	})
}

func sendSimpleErrResponse(w http.ResponseWriter, statusCode int, message string) {
	sendResp(w, statusCode, Response{
		Ok:            false,
		Message:       message,
		DataJsonBytes: nil,
	})
}

func sendSimpleBadRequestResponse(w http.ResponseWriter, message string) {
	sendResp(w, http.StatusBadRequest, Response{
		Ok:            false,
		Message:       message,
		DataJsonBytes: nil,
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
