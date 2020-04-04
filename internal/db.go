package internal

// use ORM for postgres
// https://github.com/go-pg/pg

type BuddyDb interface {
	AllUsers() []*User
	SaveUser(user *User) error
	GetUser(username string) (*User, error)
}
