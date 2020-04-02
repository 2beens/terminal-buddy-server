package internal

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
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
		w.Write([]byte("no user: " + username))
		return
	}

	userJsonData, err := json.Marshal(user)
	w.Write(userJsonData)
}

func (handler *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	username := r.FormValue("username")
	if len(username) == 0 {
		w.Write([]byte("no username"))
		return
	}

	password := r.FormValue("password")
	if len(password) == 0 {
		w.Write([]byte("no username"))
		return
	}

	passwordHashed, err := HashPassword(password)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	user := NewUser(username, passwordHashed)

	if err := handler.db.SaveUser(user); err == nil {
		w.Write([]byte("ok"))
	} else {
		w.Write([]byte(err.Error()))
	}
}
