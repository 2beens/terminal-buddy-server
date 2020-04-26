package internal

import "errors"

// use ORM for postgres
// https://github.com/go-pg/pg

type BuddyDbType int

const (
	InMemDB BuddyDbType = iota
	PsDB
)

var errorUserNotFound = errors.New("user not found")

type BuddyDb interface {
	DbOk() bool
	Close() error

	AllUsers() []*User
	SaveUser(user *User) error
	GetUser(username string) (*User, error)
	AckReminder(reminderId int64, ack bool) error
	SaveReminder(reminder *Reminder) error
	NewReminder(username string, message string, dueDate int64) error
}
