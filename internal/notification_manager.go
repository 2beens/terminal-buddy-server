package internal

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Signal struct{}

var EmptySignal = Signal{}

type NotificationManager struct {
	db           BuddyDb
	stopWorkChan chan Signal
	wsClients    map[*websocket.Conn]bool
}

func NewNotificationManager(db BuddyDb) *NotificationManager {
	nm := &NotificationManager{
		db:           db,
		stopWorkChan: make(chan Signal, 1),
		wsClients:    make(map[*websocket.Conn]bool),
	}

	go func() {
		for {
			select {
			case <-time.After(1 * time.Minute):
				if len(nm.wsClients) > 0 {
					log.Warnf("scanning %d clients for dead ws connections ...", len(nm.wsClients))
				}

				for c, _ := range nm.wsClients {
					_ = c
					// TODO:
					//if closeErr := connClient.Close(); closeErr != nil {
					//	log.Errorf("cannot close ws client conn: %s", closeErr.Error())
					//}
					//delete(rm.wsClients, connClient)
				}
			}
		}
	}()

	return nm
}

func (rm *NotificationManager) NewClient(connClient *websocket.Conn) {
	rm.wsClients[connClient] = true
	log.Debugf("notification manager got new client, total: %d", len(rm.wsClients))

	go func() {
		for {
			err := connClient.WriteMessage(websocket.TextMessage, []byte("hi from TB server ;)"))
			if err != nil {
				log.Errorf("failed to send init message to client %s: %s", connClient.RemoteAddr())
			}

			log.Tracef("waiting for messages from conn client: %s", connClient.RemoteAddr())
			msgType, message, err := connClient.ReadMessage()
			if err != nil {
				log.Errorf("notification manager read message error: %s", err.Error())
				break
			}

			log.Printf("notification manager received [type %d]: %s", msgType, message)

			err = connClient.WriteMessage(msgType, message)
			if err != nil {
				log.Printf("notification manager write error: %s", err.Error())
				break
			}
		}
	}()
}

func (rm *NotificationManager) Start() {
	for {
		select {
		case <-rm.stopWorkChan:
			log.Println("stopping reminders scan")
			return
		case <-time.After(time.Minute):
			allUsers := rm.db.AllUsers()
			//log.Printf("will scan for reminder notifications for %d users ...", len(allUsers))
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
	//log.Tracef("scanning %d reminders for user: %s", len(user.Reminders), user.Username)
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
