package internal

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Signal struct{}

var EmptySignal = Signal{}

type NotificationManager struct {
	db           BuddyDb
	stopWorkChan chan Signal
	wsClients    map[*websocket.Conn]bool
	pongWait     time.Duration // time allowed to read the next pong message from the client
}

func NewNotificationManager(db BuddyDb) *NotificationManager {
	nm := &NotificationManager{
		db:           db,
		stopWorkChan: make(chan Signal, 1),
		wsClients:    make(map[*websocket.Conn]bool),
		pongWait:     60 * time.Second,
	}

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ticker.C:
				if len(nm.wsClients) > 0 {
					log.Warnf("scanning %d clients for dead ws connections ...", len(nm.wsClients))
				}
				for c, _ := range nm.wsClients {
					if err := c.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
						log.Errorf("failed to SetWriteDeadline: %s", err.Error())
						log.Warnf("closing client conn %s", c.RemoteAddr())
						delete(nm.wsClients, c)
					}
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Errorf("failed to write ping message: %s", err.Error())
						log.Warnf("closing client conn %s", c.RemoteAddr())
						delete(nm.wsClients, c)
					}
				}
			}
		}
	}()

	return nm
}

func (rm *NotificationManager) NewClient(connClient *websocket.Conn) {
	rm.wsClients[connClient] = true
	log.Debugf("notification manager got new client, total: %d", len(rm.wsClients))

	connClient.SetPongHandler(func(string) error {
		if err := connClient.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Errorf("failed to SetReadDeadline: %s", err.Error())
		}
		return nil
	})

	err := connClient.WriteMessage(websocket.TextMessage, []byte("hi from TB server ;)"))
	if err != nil {
		log.Errorf("failed to send init message to client %s: %s", connClient.RemoteAddr())
	}

	go func() {
		for {
			log.Tracef("waiting for messages from conn client: %s", connClient.RemoteAddr())
			msgType, message, err := connClient.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Tracef("client %s going away (probably)", connClient.RemoteAddr())
					delete(rm.wsClients, connClient)
					break
				}
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
