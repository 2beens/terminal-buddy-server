package internal

type Reminder struct {
	Id      int64  `json:"id"`
	UserId  int64  `json:"-"`
	Message string `json:"message" pg:",notnull"`
	DueDate int64  `json:"due_date" pg:",notnull"`
	Ack     bool   `json:"-" pg:"default:false"` //reminder acknowledged
}

type ReminderMessage struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

type AgentMessage struct {
	UserCredentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"userCredentials"`
	Message    string `json:"message"`
	ReminderId int64  `json:"reminderId"`
}
