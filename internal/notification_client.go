package internal

import "github.com/gorilla/websocket"

type NotificationClient struct {
	User *User
	// one user can have only one device connected for now
	WsConn *websocket.Conn
}

type InitWsConnectionData struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}
