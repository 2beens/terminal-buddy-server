package internal

type Reminder struct {
	Id      int64  `json:"id"`
	UserId  int64  `json:"-"`
	Message string `json:"message" pg:",notnull"`
	DueDate int64  `json:"due_date" pg:",notnull"`
}
