package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

func sendResp(w io.Writer, response Response) {
	// TODO: pass status code

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

func logReqBody(r *http.Request) {
	buf, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		log.Print("bodyErr ", bodyErr.Error())
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	log.Printf("BODY: %q", rdr1)
	r.Body = rdr2
}

func (s *Server) getLoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAgent := r.Header.Get("User-Agent")
			log.Tracef(" ====> request [%s] path: [%s] [UA: %s]", r.Method, r.URL.Path, userAgent)
			//logReqBody(r)
			next.ServeHTTP(w, r)
		})
	}
}
