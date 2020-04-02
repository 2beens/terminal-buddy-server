package internal

type User struct {
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Reminders    []Reminder `json:"reminders"`
}

func NewUser(username, passHashed string) *User {
	return &User{
		Username:     username,
		PasswordHash: passHashed,
		Reminders:    []Reminder{},
	}
}
