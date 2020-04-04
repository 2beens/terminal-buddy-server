package internal

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type RemindHandler struct {
	db     BuddyDb
	router *mux.Router
}

func NewRemindHandler(db BuddyDb, remindRouter *mux.Router) {
	handler := &RemindHandler{
		db:     db,
		router: remindRouter,
	}

	remindRouter.HandleFunc("/{username}", handler.handleGet).Methods("GET")
	remindRouter.HandleFunc("/{username}", handler.handleNew).Methods("POST")
	remindRouter.HandleFunc("/{username}/all", handler.handleAll).Methods("GET")
	remindRouter.HandleFunc("/{username}/today", handler.handleToday).Methods("GET")
}

func (handler *RemindHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if len(username) == 0 || err != nil {
		sendSimpleErrResponse(w, "username missing / cannot get user")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, "parsing error")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, "password hash missing")
		return
	}

	if user.PasswordHash != passwordHash {
		sendSimpleErrResponse(w, "wrong credentials")
		return
	}

	rid := r.FormValue("remind_id")
	if len(rid) == 0 {
		sendSimpleErrResponse(w, "id not provided")
		return
	}

	reminder := user.GetReminder(rid)
	if reminder == nil {
		sendSimpleErrResponse(w, "not found")
		return
	}

	sendResp(w, Response{
		Ok:      true,
		Message: "ok",
		Data:    reminder,
	})
}

func (handler *RemindHandler) handleNew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if len(username) == 0 || err != nil {
		sendSimpleErrResponse(w, "username missing / cannot get user")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, "parsing error")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, "password hash missing")
		return
	}

	if user.PasswordHash != passwordHash {
		sendSimpleErrResponse(w, "wrong credentials")
		return
	}

	message := r.FormValue("message")
	dueDateStr := r.FormValue("due_date")
	if len(message) == 0 || len(dueDateStr) == 0 {
		sendSimpleErrResponse(w, "wrong arguments")
		return
	}

	dueDate, err := strconv.ParseInt(dueDateStr, 10, 64)
	if err != nil {
		sendSimpleErrResponse(w, "due date error")
		return
	}

	if err = handler.db.NewReminder(username, message, dueDate); err != nil {
		sendSimpleErrResponse(w, err.Error())
		return
	}

	sendSimpleResponse(w, "added")
}

func (handler *RemindHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if len(username) == 0 || err != nil {
		sendSimpleErrResponse(w, "username missing / cannot get user")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, "parsing error")
		return
	}

	passwordHash := r.FormValue("password_hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, "password hash missing")
		return
	}

	if user.PasswordHash != passwordHash {
		sendSimpleErrResponse(w, "wrong credentials")
		return
	}

	sendResp(w, Response{
		Ok:      true,
		Message: "ok",
		Data:    user.Reminders,
	})
}

func (handler *RemindHandler) handleToday(w http.ResponseWriter, r *http.Request) {

}
