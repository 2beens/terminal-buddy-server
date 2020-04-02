package internal

import "time"

type Reminder struct {
	Message string    `json:"message"`
	DueDate time.Time `json:"date"`
}

func NewReminder(message string, dueDate time.Time) Reminder {
	// TODO: param checks
	return Reminder{
		Message: message,
		DueDate: dueDate,
	}
}
