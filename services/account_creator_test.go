package services_test

import (
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
)

func TestAccountCreatorSuccess(t *testing.T) {
	store := mock.NewAccountStore()

	var testTable = []struct {
		config   config.Config
		username string
		password string
	}{
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "userName", "PASSword"},
		{config.Config{UsernameIsEmail: true}, "username@test.com", "PASSword"},
		{config.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "username@rightdomain.com", "PASSword"},
	}

	for _, tt := range testTable {
		acc, errs := services.AccountCreator(store, &tt.config, tt.username, tt.password)
		if len(errs) > 0 {
			for _, err := range errs {
				assert.NoError(t, err)
			}
		} else {
			assert.NotEqual(t, 0, acc.Id)
			assert.Equal(t, tt.username, acc.Username)
		}
	}
}

var pw = []byte("$2a$04$ZOBA8E3nT68/ArE6NDnzfezGWEgM6YrE17PrOtSjT5.U/ZGoxyh7e")

func TestAccountCreatorFailure(t *testing.T) {
	store := mock.NewAccountStore()
	store.Create("existing@test.com", pw)

	var testTable = []struct {
		config   config.Config
		username string
		password string
		errors   []services.Error
	}{
		// username validations
		{config.Config{}, "", "PASSword", []services.Error{{"username", "MISSING"}}},
		{config.Config{}, "  ", "PASSword", []services.Error{{"username", "MISSING"}}},
		{config.Config{}, "existing@test.com", "PASSword", []services.Error{{"username", "TAKEN"}}},
		{config.Config{UsernameIsEmail: true}, "notanemail", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "email@wrongdomain.com", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "short", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		// password validations
		{config.Config{}, "username", "", []services.Error{{"password", "MISSING"}}},
		{config.Config{PasswordMinComplexity: 2}, "username", "qwerty", []services.Error{{"password", "INSECURE"}}},
	}

	for _, tt := range testTable {
		acc, errs := services.AccountCreator(store, &tt.config, tt.username, tt.password)
		assert.Empty(t, acc)
		assert.Equal(t, tt.errors, errs)
	}
}
