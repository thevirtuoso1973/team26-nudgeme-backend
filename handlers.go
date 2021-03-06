package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// checks if user exists
func handleCheckUser(db DataSource) func(echo.Context) error {
	return func(c echo.Context) error {
		user := new(User)
		// bind the parameters into the User object
		if err := c.Bind(user); err != nil {
			return err
		}

		exists, err := db.DoesUserExist(user.Identifier)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK,
			map[string]bool{"success": true, "exists": exists})
	}
}

// adds a user to the database if the identifier is unused
func handleAddUser(db DataSource) func(echo.Context) error {
	return func(c echo.Context) error {
		user := new(User)
		if err := c.Bind(user); err != nil {
			return err
		}

		// ensure identifier is not already in use
		exists, err := db.DoesUserExist(user.Identifier)
		if err != nil {
			return err
		} else if exists {
			return failStatus(c, "Identifier already exists.")
		}

		// hash plaintext password (which is secure thanks to HTTPS)
		digest, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		err = db.InsertUser(user.Identifier, digest)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]bool{"success": true})
	}
}

// handles request to submit data to another user.
//
// If overwrite is false, it will not overwrite data between User A and User B.
func handleNewMessage(db DataSource, tableName string, overwrite bool) func(echo.Context) error {
	return func(c echo.Context) error {
		newMessage := new(NewMessageJSON)
		if err := c.Bind(newMessage); err != nil {
			return err
		}

		valid, err := db.isValidPassword(newMessage.Identifier_from, newMessage.Password)
		if err != nil {
			return err
		} else if !valid {
			return failStatus(c, "Password doesn't match expected.")
		}

		isPending, err := db.IsMessagePending(tableName, newMessage.Identifier_from,
			newMessage.Identifier_to)
		if err != nil {
			return err
		}

		toAdd, err := json.Marshal(newMessage.Data)
		if err != nil {
			return err
		}

		err = db.AddMessage(tableName, newMessage.Identifier_from, newMessage.Identifier_to,
			string(toAdd), overwrite && isPending)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]bool{"success": true})
	}
}

// handles a request to get unread messages for a given user
func handleGetMessage(db DataSource, tableName string) func(echo.Context) error {
	return func(c echo.Context) error {
		user := new(User)
		if err := c.Bind(user); err != nil {
			return err
		}

		valid, err := db.isValidPassword(user.Identifier, user.Password)
		if err != nil {
			return err
		} else if !valid {
			return failStatus(c, "Password doesn't match expected.")
		}

		messages, err := db.GetMessages(tableName, user.Identifier)
		if err != nil {
			return err
		}

		err = db.DeleteMessages(tableName, user.Identifier)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, messages)
	}
}

// data used in add_friend.html
type AddFriendTemplate struct {
	Identifier string
	PubKey     string
}

func handleAddFriend(c echo.Context) error {
	identifier := c.QueryParam("identifier")
	pubKey := c.QueryParam("pubKey")

	if isValidIDAndKey(identifier, pubKey) {
		return c.Render(http.StatusOK, "add_friend.html", AddFriendTemplate{
			Identifier: identifier,
			PubKey:     pubKey,
		})
	}
	return c.String(http.StatusBadRequest, "That link doesn't look right.")
}

// returns true if both id and key are valid
func isValidIDAndKey(id string, key string) bool {
	if len(id) == 0 || len(key) == 0 {
		return false
	}
	if len(key) < 62 {
		return false
	}
	return strings.HasPrefix(key, "-----BEGIN RSA PUBLIC KEY-----") &&
		strings.HasSuffix(key, "-----END RSA PUBLIC KEY-----")
}

// returns true if given password matches the password linked with identifier in DB
func verifyIdentity(db DataSource, identifier string, password string) bool {
	isValid, err := db.isValidPassword(identifier, password)
	println(err)
	return isValid
}

func failStatus(c echo.Context, reason string) error {
	return c.JSON(http.StatusBadRequest,
		map[string]interface{}{"success": false, "reason": reason})
}
