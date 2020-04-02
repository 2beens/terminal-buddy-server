package internal

type BuddyDb interface {
	AllUsers() []*User
	SaveUser(user *User) error
	GetUser(username string) (*User, error)
}
