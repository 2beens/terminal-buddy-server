package internal

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "username missing / cannot get user")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, "parsing error")
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

	rid := r.FormValue("remind_id")
	if len(rid) == 0 {
		sendSimpleBadRequestResponse(w, "id not provided")
		return
	}

	reminder := user.GetReminder(rid)
	if reminder == nil {
		sendSimpleErrResponse(w, http.StatusNotFound, "not found")
		return
	}

	sendResp(w, http.StatusOK, Response{
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
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "username missing / cannot get user")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		sendSimpleErrResponse(w, http.StatusInternalServerError, "parsing error")
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

	message := r.FormValue("message")
	dueDateStr := r.FormValue("due_date")
	if len(message) == 0 || len(dueDateStr) == 0 {
		sendSimpleBadRequestResponse(w, "wrong arguments")
		return
	}

	log.Println("message: " + message)
	log.Println("dueDateStr: " + dueDateStr)

	dueDate, err := strconv.ParseInt(dueDateStr, 10, 64)
	if err != nil {
		sendSimpleErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("due date (%v) error", dueDateStr))
		return
	}

	if err = handler.db.NewReminder(username, message, dueDate); err != nil {
		sendSimpleErrResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSimpleResponse(w, "added")
}

func (handler *RemindHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if len(username) == 0 || err != nil {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "username missing / cannot get user")
		return
	}

	passwordHash := r.Header.Get("Term-Buddy-Pass-Hash")
	if len(passwordHash) == 0 {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "password hash missing")
		return
	}

	if user.PasswordHash != passwordHash {
		sendSimpleErrResponse(w, http.StatusNotAcceptable, "wrong credentials")
		return
	}

	sendResp(w, http.StatusOK, Response{
		Ok:      true,
		Message: "ok",
		Data:    user.Reminders,
	})
}

func (handler *RemindHandler) handleToday(w http.ResponseWriter, r *http.Request) {

}
