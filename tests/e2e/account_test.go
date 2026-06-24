package main

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stepRegister(t *testing.T) {
	query := `
        mutation Register($account: RegisterInput!) {
          register(account: $account) {
            token
          }
        }
    `
	Email = fmt.Sprintf("random%d@example.com", rand.Intn(100000))
	Password = fmt.Sprintf("password%d", rand.Intn(100000))
	variables := map[string]interface{}{
		"account": map[string]interface{}{
			"name":     "John Doe",
			"email":    Email,
			"password": Password,
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during Register")

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "response Data should be a map")

	reg, ok := data["register"].(map[string]interface{})
	assert.True(t, ok, "register field should be a map")

	token, ok := reg["token"].(string)
	assert.True(t, ok, "token should be a string")
	assert.NotEmpty(t, token, "expected a token in register response")

	AuthToken = token
	setAccountIDFromToken(t)
}

func stepLogin(t *testing.T) {
	query := `
        mutation Login($account: LoginInput!) {
          login(account: $account) {
            token
          }
        }
    `
	variables := map[string]interface{}{
		"account": map[string]interface{}{
			"email":    Email,
			"password": Password,
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during Login")

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "response Data should be a map")

	login, ok := data["login"].(map[string]interface{})
	assert.True(t, ok, "login field should be a map")

	token, ok := login["token"].(string)
	assert.True(t, ok, "token should be a string")
	assert.NotEmpty(t, token, "expected a token in login response")

	AuthToken = token
	log.Println("Got token from Login")
}

func stepQueryAccounts(t *testing.T) {
	query := `
        query GetAccounts($pagination: PaginationInput, $id: Int) {
          accounts(pagination: $pagination, id: $id) {
            id
            name
            email
          }
        }
    `
	variables := map[string]interface{}{
		"pagination": map[string]interface{}{
			"skip": 0,
			"take": 10,
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	accounts, ok := data["accounts"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, accounts)
	log.Println("Accounts:", accounts)
}
