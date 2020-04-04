package internal

import (
	"errors"
)

var errorUserNotFound = errors.New("user not found")

// TODO: solve multi thread problems

type MemDb struct {
	users map[string]*User
}

func NewMemDb() *MemDb {
	return &MemDb{
		users: make(map[string]*User),
	}
}

func (db *MemDb) AllUsers() []*User {
	var allUsers []*User
	for _, u := range db.users {
		allUsers = append(allUsers, u)
	}
	return allUsers
}

func (db *MemDb) SaveUser(user *User) error {
	db.users[user.Username] = user
	return nil
}

func (db *MemDb) GetUser(username string) (*User, error) {
	for u, _ := range db.users {
		if username == u {
			return db.users[u], nil
		}
	}
	return nil, errorUserNotFound
}

func (db *MemDb) NewReminder(username string, message string, dueDate int64) error {
	user, err := db.GetUser(username)
	if err != nil {
		return err
	}

	id := string(len(user.Reminders))
	reminder := NewReminder(id, message, dueDate)
	user.Reminders = append(user.Reminders, &reminder)

	return nil
}
