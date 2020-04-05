package internal

type Reminder struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
	DueDate int64  `json:"due_date"`
}

func NewReminder(id int, message string, dueDate int64) Reminder {
	return Reminder{
		Id:      id,
		Message: message,
		DueDate: dueDate,
	}
}
