package internal

type Reminder struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	DueDate int64  `json:"due_date"`
}

func NewReminder(id, message string, dueDate int64) Reminder {
	return Reminder{
		Id:      id,
		Message: message,
		DueDate: dueDate,
	}
}
