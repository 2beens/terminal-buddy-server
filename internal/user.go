package internal

type User struct {
	Id           int64       `json:"-"`
	Username     string      `json:"username" pg:",unique,notnull"`
	PasswordHash string      `json:"-"`
	Reminders    []*Reminder `json:"reminders" pg:"-"`
	// TODO: maybe add email address and a requirement to verify
}

func (u *User) GetReminder(id int64) *Reminder {
	for _, r := range u.Reminders {
		if r.Id == id {
			return r
		}
	}
	return nil
}
