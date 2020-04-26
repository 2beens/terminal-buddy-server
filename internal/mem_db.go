package internal

import (
	"errors"
	"fmt"
	"time"
)

// TODO: solve multi thread problems

type MemDb struct {
	users         map[int64]*User
	reminder2user map[int64]int64
}

func (db *MemDb) DbOk() bool {
	return true
}

func (db *MemDb) Close() error {
	return nil
}

func NewMemDb() *MemDb {
	return &MemDb{
		users:         make(map[int64]*User),
		reminder2user: make(map[int64]int64),
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
	db.users[user.Id] = user
	return nil
}

func (db *MemDb) GetUser(username string) (*User, error) {
	for id, _ := range db.users {
		if db.users[id].Username == username {
			return db.users[id], nil
		}
	}
	return nil, errorUserNotFound
}

func (db *MemDb) AckReminder(reminderId int64, ack bool) error {
	userId, ok := db.reminder2user[reminderId]
	if !ok {
		return fmt.Errorf("cannot find coresponding user")
	}

	user, ok := db.users[userId]
	if !ok {
		return fmt.Errorf("cannot find coresponding user")
	}

	reminder, err := db.getReminder(user.Id, reminderId)
	if err != nil {
		return err
	}

	reminder.Ack = true

	return nil
}

func (db *MemDb) getReminder(userId, reminderId int64) (*Reminder, error) {
	user, ok := db.users[userId]
	if !ok {
		return nil, errors.New("cannot find user")
	}

	for i, _ := range user.Reminders {
		if user.Reminders[i].Id == reminderId {
			return user.Reminders[i], nil
		}
	}

	return nil, errors.New("not found")
}

func (db *MemDb) SaveReminder(reminder *Reminder) error {
	foundReminder, err := db.getReminder(reminder.UserId, reminder.Id)
	if err != nil {
		return err
	}

	foundReminder.Ack = reminder.Ack
	foundReminder.Message = reminder.Message
	foundReminder.DueDate = reminder.DueDate

	return nil
}

func (db *MemDb) NewReminder(username string, message string, dueDate int64) error {
	user, err := db.GetUser(username)
	if err != nil {
		return err
	}

	reminderId := time.Now().Unix()

	db.reminder2user[reminderId] = user.Id

	user.Reminders = append(user.Reminders, &Reminder{
		Id:      reminderId,
		Message: message,
		DueDate: dueDate,
	})

	return nil
}
