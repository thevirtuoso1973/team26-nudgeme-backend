package main

import (
	"database/sql"
	"encoding/json"

	"golang.org/x/crypto/bcrypt"
)

// interface that defines the methods needed to interact with database
// for wellbeing sharing
type DataSource interface {
	DoesUserExist(identifier string) (bool, error)

	// @param password is plaintext
	isValidPassword(identifier string, password string) (bool, error)

	// inserts the identifier and hashed password digest
	InsertUser(identifier string, digest []byte) error

	// returns true if there is an existing message pending to be sent
	// between users
	IsMessagePending(tableName string, identifier_from string, identifier_to string) (bool, error)

	// updates if overwrite is true else inserts a new row.
	// Doesn't check if appropriate row exists in the first place, so check that
	// if overwriting.
	AddMessage(tableName string, identifier_from string, identifier_to string,
		data string, overwrite bool) error

	// gets the list of messages sent to this user
	GetMessages(tableName string, identifier string) ([]interface{}, error)

	// deletes the messages sent to this user
	DeleteMessages(tableName string, identifier string) error
}

// new type since we can't implement extensions to the sql.DB type
type MyDB struct {
	database *sql.DB
}

func (mydb *MyDB) DoesUserExist(identifier string) (bool, error) {
	sqlDB := mydb.database

	count := 0
	err := sqlDB.QueryRow("SELECT COUNT(*) FROM users WHERE identifier = ?",
		identifier).Scan(&count)

	return count > 0, err
}

func (mydb *MyDB) InsertUser(identifier string, digest []byte) error {
	db := mydb.database

	_, err := db.Exec("INSERT INTO users (identifier, password) VALUES (?, ?)",
		identifier, digest)
	return err
}

func (mydb *MyDB) IsMessagePending(tableName string,
	identifier_from string, identifier_to string) (bool, error) {
	db := mydb.database

	count := 0
	countQuery := "SELECT COUNT(*) FROM " + tableName + " WHERE " +
		"identifier_from = ? AND identifier_to = ?"
	err := db.QueryRow(countQuery,
		identifier_from, identifier_to).Scan(&count)

	return count > 0, err
}

func (mydb *MyDB) AddMessage(tableName string,
	identifier_from string, identifier_to string,
	data string, overwrite bool) error {
	db := mydb.database

	var err error
	if overwrite {
		updateQuery := "UPDATE " + tableName + " SET data = ? " +
			"WHERE identifier_from = ? AND identifier_to = ?"
		_, err = db.Exec(updateQuery,
			data, identifier_from, identifier_to)
	} else {
		insertQuery := "INSERT INTO " + tableName + " (identifier_from, " +
			"identifier_to, data) VALUES (?, ?, ?)"
		_, err = db.Exec(insertQuery,
			identifier_from, identifier_to, data)
	}
	return err
}

func (mydb *MyDB) isValidPassword(identifier string, password string) (bool, error) {
	db := mydb.database

	var stored []byte
	err := db.QueryRow("SELECT password FROM users WHERE identifier = ? LIMIT 1",
		identifier).Scan(&stored)
	if err == nil {
		err := bcrypt.CompareHashAndPassword(stored, []byte(password))
		return err == nil, err // valid if no error
	} else {
		return false, err
	}
}

func (mydb *MyDB) GetMessages(tableName string, identifier string) ([]interface{}, error) {
	db := mydb.database

	query := "SELECT identifier_from, data FROM " + tableName + " WHERE identifier_to = ?"
	rows, err := db.Query(query, identifier)

	messages := make([]interface{}, 0)
	for rows.Next() {
		var identifier_from string
		var encoded []byte
		var decoded interface{}

		rows.Scan(&identifier_from, &encoded)
		// query seems to return json strings so I decode here; we may
		// as well send actual JSON
		json.Unmarshal(encoded, &decoded)

		messages = append(messages,
			map[string]interface{}{"identifier_from": identifier_from, "data": decoded})
	}

	return messages, err
}

func (mydb *MyDB) DeleteMessages(tableName string, identifier string) error {
	db := mydb.database

	queryDelete := "DELETE FROM " + tableName + " WHERE identifier_to = ?"
	_, err := db.Exec(queryDelete, identifier)

	return err
}
