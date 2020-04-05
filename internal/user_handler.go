package internal

import (
	"encoding/json"
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

	userRouter.HandleFunc("/login", handler.handleLogin).Methods("POST")
	userRouter.HandleFunc("/register", handler.handleRegister).Methods("POST")
}

func (handler *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, "parsing error")
		return
	}

	username := r.FormValue("username")
	user, err := handler.db.GetUser(username)
	if len(username) == 0 || err != nil {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "username missing / cannot get user")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "password hash missing")
		return
	}

	if user.PasswordHash != passwordHash {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "wrong credentials")
		return
	}

	userJsonBytes, err := json.Marshal(user)
	if err != nil {
		log.Errorf("error marshaling user [%s]: %s", user.Username, err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, "marshaling error")
		return
	}

	sendResp(w, http.StatusOK, Response{
		Ok:            true,
		Message:       "ok",
		DataJsonBytes: userJsonBytes,
	})
}

func (handler *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, "parsing error")
		return
	}

	username := r.FormValue("username")
	if len(username) == 0 {
		sendSimpleBadRequestResponse(w, "username missing")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleBadRequestResponse(w, "password hash missing")
		return
	}

	user := NewUser(username, passwordHash)

	if err := handler.db.SaveUser(user); err == nil {
		sendSimpleResponse(w, "ok")
	} else {
		log.Errorf("error saving new user: %s", err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, err.Error())
	}
}
