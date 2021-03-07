package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mocked object that implements DataSource
type FakeDB struct {
	mock.Mock
}

func TestCheckUserExists(t *testing.T) {
	// set up the values and mocked object:
	identifier := "existing"
	body := "{\"identifier\":\"" + identifier + "\"}"
	fakeDB := new(FakeDB)
	fakeDB.On("DoesUserExist", identifier).Return(true, nil)

	// set up the fake request, and a recorder
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// perform the assertions, verify assumptions
	if assert.NoError(t, handleCheckUser(fakeDB)(c)) {
		fakeDB.AssertExpectations(t)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "\"exists\":true")
	}
}

func TestCheckUserDoesntExists(t *testing.T) {
	identifier := "not-existing"
	body := "{\"identifier\":\"" + identifier + "\"}"
	fakeDB := new(FakeDB)
	fakeDB.On("DoesUserExist", identifier).Return(false, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, handleCheckUser(fakeDB)(c)) {
		fakeDB.AssertExpectations(t)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "\"exists\":false")
	}
}

func TestAddUser(t *testing.T) {
	identifier := "user"
	password := "battery horse staple"
	body := "{\"identifier\":\"" + identifier + "\", \"password\": \""+ password +"\"}"

	fakeDB := new(FakeDB)
	fakeDB.On("DoesUserExist", identifier).Return(false, nil)
	// NOTE: cannot easily compute the hash digest of password since bcrypt
	// uses salts, therefore only verifying type
	fakeDB.On("InsertUser", identifier, mock.AnythingOfType("[]uint8")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/user/new", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	if assert.NoError(t, handleAddUser(fakeDB)(c)) {
		fakeDB.AssertExpectations(t)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "\"success\":true")
	}
}

func TestDoesNotAddExistingUser(t *testing.T) {
	identifier := "user"
	password := "battery horse staple"
	body := "{\"identifier\":\"" + identifier + "\", \"password\": \""+ password +"\"}"

	fakeDB := new(FakeDB)
	fakeDB.On("DoesUserExist", identifier).Return(true, nil)

	req := httptest.NewRequest(http.MethodPost, "/user/new", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	if assert.NoError(t, handleAddUser(fakeDB)(c)) {
		fakeDB.AssertExpectations(t)
		fakeDB.AssertNotCalled(t, "InsertUser", identifier, mock.AnythingOfType("[]uint8"))

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "\"success\":false")
	}
}

func (db *FakeDB) DoesUserExist(identifier string) (bool, error) {
	args := db.Called(identifier)
	// these behave as strongly typed getters
	return args.Bool(0), args.Error(1)
}

func (mydb *FakeDB) InsertUser(identifier string, digest []byte) error {
	args := mydb.Called(identifier, digest)
	return args.Error(0)
}

func (mydb *FakeDB) IsMessagePending(
	tableName string,
	identifier_from string, identifier_to string) (bool, error) {
	args := mydb.Called(tableName, identifier_from, identifier_to)
	return args.Bool(0), args.Error(1)
}

func (mydb *FakeDB) AddMessage(tableName string, identifier_from string, identifier_to string,
	data string, wasPending bool) error {
	args := mydb.Called(tableName, identifier_from, identifier_to, data, wasPending)
	return args.Error(0)
}

func (mydb *FakeDB) isValidPassword(identifier string, password string) (bool, error) {
	args := mydb.Called(identifier, password)
	return args.Bool(0), args.Error(1)
}

func (mydb *FakeDB) GetMessages(tableName string, identifier string) ([]interface{}, error) {
	args := mydb.Called(tableName, identifier)

	// we'll panic if first arg is not the expected type
	return args.Get(0).([]interface{}), args.Error(1)
}

func (mydb *FakeDB) DeleteMessages(tableName string, identifier string) error {
	args := mydb.Called(tableName, identifier)
	return args.Error(0)
}
