package internal

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Signal struct{}

var EmptySignal = Signal{}

type NotificationManager struct {
	db           BuddyDb
	stopWorkChan chan Signal
}

func NewNotificationManager(db BuddyDb) *NotificationManager {
	return &NotificationManager{
		db:           db,
		stopWorkChan: make(chan Signal, 1),
	}
}

func (rm *NotificationManager) Start() {
	for {
		select {
		case <-rm.stopWorkChan:
			log.Println("stopping reminders scan")
			return
		case <-time.After(time.Minute):
			allUsers := rm.db.AllUsers()
			log.Printf("will scan for reminder notifications for %d users ...", len(allUsers))
			rm.ScanRemindersForUsers(allUsers)
		}
	}
}

func (rm *NotificationManager) Stop() {
	rm.stopWorkChan <- EmptySignal
}

func (rm *NotificationManager) ScanRemindersForUsers(users []*User) {
	now := time.Now().Truncate(time.Minute)
	for _, user := range users {
		rm.scanNotificationsForUser(now, user)
	}
}

func (rm *NotificationManager) scanNotificationsForUser(now time.Time, user *User) {
	log.Tracef("scanning %d reminders for user: %s", len(user.Reminders), user.Username)
	for _, reminder := range user.Reminders {
		dueDate := time.Unix(reminder.DueDate, 0).Truncate(time.Minute)
		if now.Equal(dueDate) {
			rm.sendNotification(user, reminder)
		}
	}
}

func (rm *NotificationManager) sendNotification(user *User, reminder *Reminder) {
	log.Tracef("sending notification (%s) to user %s", reminder.Message, user.Username)
	// TODO:
}
