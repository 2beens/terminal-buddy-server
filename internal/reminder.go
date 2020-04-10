package internal

type Reminder struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
	DueDate int64  `json:"due_date"`
}
