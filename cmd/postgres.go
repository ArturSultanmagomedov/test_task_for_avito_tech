package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
)

const schema = `
create table if not exists users
(
    id      serial primary key,
    user_id int   not null unique,
    balance float not null
);`

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(c Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.DBName, c.Password, c.SSLMode))
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Database struct {
	db *sqlx.DB
}

func NewDatabase(db *sqlx.DB) *Database {
	return &Database{db: db}
}

func (r Database) CreateUsersTable() error {
	_, err := r.db.Query(schema)
	if err != nil {
		return err
	}
	return nil
}

func (r Database) CreateUser(userId int, balance float32) error {
	_, err := r.db.Query("insert into users (user_id, balance) values ($1, $2);", userId, balance)
	if err != nil {
		return err
	}
	return nil
}

func (r Database) GetUser(userId int) (*User, error) {
	var user User
	err := r.db.Get(&user, "select * from users where user_id = $1", userId)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r Database) IsUserExist(userId int) (bool, error) {
	var c int
	err := r.db.Get(&c, "select count(1) from users where user_id = $1;", userId)
	if err != nil {
		return false, err
	}

	return c > 0, nil
}

func (r Database) UpdateBalance(userId int, sum float32) (*User, error) {
	var user User
	err := r.db.Get(&user, "select * from users where user_id = $1;", userId)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Query("update users set balance = $1 where user_id = $2;", user.Balance+sum, userId)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r Database) CreateFundsTransaction(id1, id2 int, sum float32) error {

	var user1 User
	var user2 User

	err := r.db.Get(&user1, "select * from users where user_id = $1;", id1)
	if err != nil {
		return err
	}

	err = r.db.Get(&user2, "select * from users where user_id = $1;", id2)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("update users set balance = $1 where user_id = $2;", user1.Balance-sum, id1)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatalf("update failed: %v, unable to back: %v", err, rollbackErr)
		}
		return err
	}

	_, err = tx.Exec("update users set balance = $1 where user_id = $2;", user2.Balance+sum, id2)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatalf("update failed: %v, unable to back: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}
