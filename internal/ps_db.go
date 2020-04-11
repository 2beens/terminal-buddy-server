package internal

import (
	"crypto/md5"
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	log "github.com/sirupsen/logrus"
)

type PostgresDBClient struct {
	db *pg.DB
}

func NewPostgresDBClient(recreateDb bool) (*PostgresDBClient, error) {
	c := &PostgresDBClient{db: pg.Connect(&pg.Options{
		ApplicationName: "terminal-buddy",
		User:            "termbuddy",
		Database:        "termbuddydb",
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
	return nil
}

func (c *PostgresDBClient) SaveUser(user *User) error {
	return nil
}

func (c *PostgresDBClient) GetUser(username string) (*User, error) {
	return nil, nil
}

func (c *PostgresDBClient) NewReminder(username string, message string, dueDate int64) error {
	return nil
}
