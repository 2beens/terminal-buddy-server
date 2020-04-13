package internal

import (
	"crypto/md5"
	"errors"
	"fmt"

	"TerminalBuddyServer/config"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	log "github.com/sirupsen/logrus"
)

type PostgresDBClient struct {
	db *pg.DB
}

func NewPostgresDBClient(config *config.TBConfig, dbPassword string, recreateDb bool) (*PostgresDBClient, error) {
	c := &PostgresDBClient{db: pg.Connect(&pg.Options{
		ApplicationName: "terminal-buddy",
		Database:        config.DB.Name,
		User:            config.DB.User,
		Password:        dbPassword,
	})}

	err := c.createSchema(recreateDb)
	if err != nil {
		return nil, err
	}

	if c.insertAdminUser() {
		log.Debug("admin user added")
	}

	return c, nil
}

func (c *PostgresDBClient) createSchema(recreateDb bool) error {
	if recreateDb {
		for _, model := range []interface{}{(*User)(nil), (*Reminder)(nil)} {
			err := c.db.DropTable(model, &orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, model := range []interface{}{(*User)(nil), (*Reminder)(nil)} {
		err := c.db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *PostgresDBClient) insertAdminUser() bool {
	admin := User{
		Username:     "serj",
		PasswordHash: fmt.Sprintf("%x", md5.Sum([]byte("serj"))),
		Reminders:    nil,
	}

	created, err := c.db.Model(&admin).
		Column("id").
		Where("username = ?username").
		OnConflict("DO NOTHING"). // OnConflict is optional
		Returning("id").
		SelectOrInsert()
	if err != nil {
		panic(err)
	}

	return created
}

func (c *PostgresDBClient) DbOk() bool {
	_, err := c.db.Exec("SELECT 1")
	if err != nil {
		return false
	}
	return true
}

func (c *PostgresDBClient) Close() error {
	return c.db.Close()
}

func (c *PostgresDBClient) AllUsers() []*User {
	var users []User
	err := c.db.Model(&users).Select()
	if err != nil {
		panic(err)
	}
	var allUsers []*User
	for _, u := range users {
		allUsers = append(allUsers, &u)
	}
	return allUsers
}

func (c *PostgresDBClient) SaveUser(user *User) error {
	res, err := c.db.Model(user).
		Returning("id").
		OnConflict("(id) DO UPDATE").
		Set("password_hash = EXCLUDED.password_hash").
		Insert()
	if err != nil {
		return err
	}
	if res.RowsAffected() <= 0 {
		return errors.New("user not saved")
	}
	return nil
}

func (c *PostgresDBClient) GetUser(username string) (*User, error) {
	user := &User{
		Username: username,
	}
	err := c.db.Model(user).
		Where("username = ?username").
		Select()
	if err != nil {
		return nil, err
	}

	userReminders, err := c.getUserReminders(user.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot get user %s reminders: %w", username, err)
	}

	user.Reminders = userReminders

	return user, nil
}

func (c *PostgresDBClient) getUserReminders(userId int64) ([]*Reminder, error) {
	var remindersFromDb []Reminder
	err := c.db.Model(&remindersFromDb).
		Where("user_id = ?", userId).
		Select()
	if err != nil {
		return nil, fmt.Errorf("cannot get reminders for user %d: %w", userId, err)
	}

	var reminders []*Reminder
	for i, _ := range remindersFromDb {
		reminders = append(reminders, &remindersFromDb[i])
	}

	return reminders, nil
}

func (c *PostgresDBClient) NewReminder(username string, message string, dueDate int64) error {
	user, err := c.GetUser(username)
	if err != nil {
		return fmt.Errorf("cannot find user %s: %w", username, err)
	}

	reminder := &Reminder{
		UserId:  user.Id,
		Message: message,
		DueDate: dueDate,
	}

	res, err := c.db.Model(reminder).
		Returning("id").
		Insert()
	if err != nil {
		return err
	}

	if res.RowsAffected() <= 0 {
		return errors.New("reminder not stored")
	}

	return nil
}
