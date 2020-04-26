package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Signal struct{}

var EmptySignal = Signal{}

type NotificationManager struct {
	db                  BuddyDb
	stopWorkChan        chan Signal
	notificationClients map[string]*NotificationClient
	pongWait            time.Duration // time allowed to read the next pong message from the client
}

func NewNotificationManager(db BuddyDb) *NotificationManager {
	nm := &NotificationManager{
		db:                  db,
		stopWorkChan:        make(chan Signal, 1),
		notificationClients: make(map[string]*NotificationClient), // username <-> conn
		pongWait:            60 * time.Second,
	}

	go nm.ScanDeadWsConnections()

	return nm
}

func (nm *NotificationManager) NewClient(connClient *websocket.Conn) {
	log.Debugf("notification manager got new client, total before: %d", len(nm.notificationClients))

	// client has to first send its init message (username, password), then we add the connection
	initData := &InitWsConnectionData{}
	_, initMessage, err := connClient.ReadMessage()
	if err != nil {
		log.Errorf("notification manager read init message error: %s", err.Error())
		connClient.Close()
		return
	}

	err = json.Unmarshal(initMessage, initData)
	if err != nil {
		log.Errorf("failed to read init data for %s", connClient.RemoteAddr())
		if err := connClient.WriteMessage(websocket.TextMessage, []byte("corrupt init data")); err != nil {
			log.Errorf("ws corrupt init file, failed to send error response to client %s: %s", initData.Username, err.Error())
		}
		connClient.Close()
		return
	}

	user, err := nm.db.GetUser(initData.Username)
	if err != nil {
		log.Errorf("ws conn failed, cannot find user %s", initData.Username)
		if err := connClient.WriteMessage(websocket.TextMessage, []byte("wrong user data")); err != nil {
			log.Errorf("ws cannot fund user, failed to send error response to client %s: %s", initData.Username, err.Error())
		}
		connClient.Close()
		return
	}

	if user.PasswordHash != initData.PasswordHash {
		log.Errorf("ws conn failed, wrong credentials for %s", initData.Username)
		if err := connClient.WriteMessage(websocket.TextMessage, []byte("wrong user data")); err != nil {
			log.Errorf("ws wrong credentials, failed to send error response to client %s: %s", initData.Username, err.Error())
		}
		connClient.Close()
		return
	}

	nc := &NotificationClient{
		User:   user,
		WsConn: connClient,
	}

	nm.notificationClients[user.Username] = nc

	connClient.SetPongHandler(func(string) error {
		//log.Tracef("sending pong to %s", connClient.RemoteAddr())
		if err := connClient.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Errorf("failed to SetReadDeadline: %s", err.Error())
		}
		return nil
	})

	connClient.SetPingHandler(func(appData string) error {
		log.Tracef("received ping: %s", appData)
		return nil
	})

	err = nc.WsConn.WriteMessage(websocket.TextMessage, []byte("hi from TB server ;)"))
	if err != nil {
		log.Errorf("failed to send init message to client %s: %s", connClient.RemoteAddr(), err.Error())
	}

	go nm.WatchWsClient(nc)
}

func (nm *NotificationManager) WatchWsClient(nc *NotificationClient) {
	for {
		log.Tracef("waiting for messages from conn client: %s", nc.WsConn.RemoteAddr())
		msgType, message, err := nc.WsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Tracef("client %s going away (probably)", nc.WsConn.RemoteAddr())
			} else {
				log.Errorf("notification manager read message error: %s", err.Error())
			}
			nm.RemoveNotificationClient(nc)
			break
		}

		log.Printf("notification manager received [type %d]: %s", msgType, message)

		// try to read agentMessage
		var agentMessage AgentMessage
		err = json.Unmarshal(message, &agentMessage)
		if err != nil {
			log.Errorf("unmarshal agentMessage error, incoming message is not agentMessage")
		} else {
			if agentMessage.Message == "ack" {
				if err := nm.db.AckReminder(agentMessage.ReminderId, true); err != nil {
					log.Errorf("failed to ACK reminder %d: %s", agentMessage.ReminderId, err)
				} else {
					log.Tracef("reminder %d ACKd", agentMessage.ReminderId)
				}
			}
			continue
		}

		echoedMessage := fmt.Sprintf("[WIP] echo: %s", message)
		err = nc.WsConn.WriteMessage(msgType, []byte(echoedMessage))
		if err != nil {
			log.Printf("notification manager write error: %s", err.Error())
			// TODO: maybe check error before killing WS connection
			nm.RemoveNotificationClient(nc)
			break
		}
	}
}

func (nm *NotificationManager) RemoveNotificationClient(nc *NotificationClient) {
	log.Warnf("removing notification client for user %s", nc.User.Username)
	delete(nm.notificationClients, nc.User.Username)
}

func (nm *NotificationManager) Start() {
	for {
		select {
		case <-nm.stopWorkChan:
			log.Println("stopping reminders scan")
			return
		case <-time.After(time.Minute):
			allUsers := nm.db.AllUsers()
			//log.Printf("will scan for reminder notifications for %d users ...", len(allUsers))
			nm.ScanRemindersForUsers(allUsers)
		}
	}
}

func (nm *NotificationManager) Stop() {
	nm.stopWorkChan <- EmptySignal
}

func (nm *NotificationManager) ScanDeadWsConnections() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		log.Warn("stopped scanning dead connections")
	}()
	for {
		select {
		case <-ticker.C:
			//if len(nm.notificationClients) > 0 {
			//	log.Tracef("scanning %d clients for dead ws connections ...", len(nm.notificationClients))
			//}
			for _, c := range nm.notificationClients {
				if err := c.WsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Errorf("failed to write ping message: %s", err.Error())
					log.Warnf("closing client conn %s", c.WsConn.RemoteAddr())
					nm.RemoveNotificationClient(c)
				}
			}
		}
	}
}

func (nm *NotificationManager) ScanRemindersForUsers(users []*User) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("ScanRemindersForUsers recovered from panic: %s", r)
		}
	}()

	now := time.Now().Truncate(time.Minute)
	for _, user := range users {
		nm.scanNotificationsForUser(now, user)
	}
}

func (nm *NotificationManager) scanNotificationsForUser(now time.Time, user *User) {
	//log.Tracef("scanning %d reminders for user: %s", len(user.Reminders), user.Username)
	for _, reminder := range user.Reminders {
		dueDate := time.Unix(reminder.DueDate, 0).Truncate(time.Minute)
		if now.Equal(dueDate) {
			nm.sendNotification(user, reminder)
		} else if !reminder.Ack && now.After(dueDate) { // reminder was not set - agent was offline
			nm.sendNotification(user, reminder)
		}
	}
}

func (nm *NotificationManager) sendNotification(user *User, reminder *Reminder) {
	log.Tracef("will try sending notification (%s) to user %s", reminder.Message, user.Username)

	reminderMessage := ReminderMessage{
		Id:      reminder.Id,
		Message: reminder.Message,
	}
	reminderMessageBytes, err := json.Marshal(reminderMessage)
	if err != nil {
		log.Errorf("marshal reminder message failed for reminder: %s", reminder.Id)
		return
	}

	nc, ok := nm.notificationClients[user.Username]
	if !ok {
		//log.Tracef("agent for user %s not connected, skip sending notification", user.Username)
		return
	}

	err = nc.WsConn.WriteMessage(websocket.TextMessage, reminderMessageBytes)
	if err != nil {
		log.Errorf("failed to send reminder message to client %s: %s", nc.WsConn.RemoteAddr(), err.Error())
	}
}
