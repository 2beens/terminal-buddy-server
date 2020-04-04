package internal

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type UserHandler struct {
	db     BuddyDb
	router *mux.Router
}

func NewUserHandler(db BuddyDb, userRouter *mux.Router) {
	handler := &UserHandler{
		db:     db,
		router: userRouter,
	}

	userRouter.HandleFunc("/{username}", handler.handleGet).Methods("GET")
	userRouter.HandleFunc("/register", handler.handleRegister).Methods("POST")
}

func (handler *UserHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if err != nil {
		sendSimpleErrResponse(w, "username missing")
		return
	}

	sendResp(w, Response{
		Ok:      true,
		Message: "ok",
		Data:    user,
	})
}

func (handler *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, "parsing error")
		return
	}

	username := r.FormValue("username")
	if len(username) == 0 {
		sendSimpleErrResponse(w, "username missing")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, "password hash missing")
		return
	}

	user := NewUser(username, passwordHash)

	if err := handler.db.SaveUser(user); err == nil {
		sendSimpleResponse(w, "ok")
	} else {
		log.Errorf("error saving new user: %s", err.Error())
		sendSimpleErrResponse(w, err.Error())
	}
}
