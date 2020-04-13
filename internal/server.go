package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"TerminalBuddyServer/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	port int
	db   BuddyDb

	wsUpgrader          websocket.Upgrader
	notificationManager *NotificationManager
}

func NewServer(tbConfig *config.TBConfig, dbType BuddyDbType, dbPassword string, recreateDb bool) *Server {
	server := &Server{
		wsUpgrader: websocket.Upgrader{}, // use default options
		port:       tbConfig.Port(),
	}

	log.Tracef("config: %v", tbConfig)

	if dbType == InMemDB {
		server.db = NewMemDb()
		log.Println("using in memory DB")
	} else if dbType == PsDB {
		var err error
		if server.db, err = NewPostgresDBClient(tbConfig, dbPassword, recreateDb); err != nil {
			panic(err)
		}
		log.Println("using Postgres DB")
	} else {
		panic("unknown DB type")
	}

	if !server.db.DbOk() {
		panic("DB connection not happy ...")
	}

	server.notificationManager = NewNotificationManager(server.db)

	return server
}

func (s *Server) Serve() {
	router := s.routerSetup()

	ipAndPort := fmt.Sprintf("%s:%d", "localhost", s.port)
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

	go func() {
		s.notificationManager.Start()
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

func (s *Server) routerSetup() *mux.Router {
	log.Trace("setting routes")
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sendSimpleResponse(w, "WIP")
	})

	r.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("new websocket client connecting: %s", r.RemoteAddr)

		c, err := s.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("WS upgrade error: %s", err.Error())
			return
		}
		defer func() {
			err := c.Close()
			if err != nil {
				log.Errorf("failed to close WS connection: %s", err.Error())
			}
		}()

		// pass client connection to notification manager
		s.notificationManager.NewClient(c)
	})

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		sendSimpleResponse(w, "i'm fine <3")
	})

	// handle register
	NewUserHandler(s.db, r.PathPrefix("/user").Subrouter())

	// handle remind
	NewRemindHandler(s.db, r.PathPrefix("/remind").Subrouter())

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
