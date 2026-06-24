package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

// GraphQLRequest is a helper struct for the request body
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse is a helper struct for the response body
type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

// Change this if needed:
var (
	serverURL = "http://localhost:8080/graphql"
	Email     string
	Password  string
	AuthToken string
)

// doRequest is a helper that executes a GraphQL mutation/query
// against our server, attaching the JWT token as a *cookie*
// if AuthToken is set.
func doRequest(t *testing.T, serverURL, query string, variables map[string]interface{}) GraphQLResponse {
	body := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	// Encode request to JSON
	b, err := json.Marshal(body)
	assert.NoError(t, err)

	// Build the request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(b))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// If we have a token, set it as a cookie named "token"
	if AuthToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "token",
			Value: AuthToken,
			Path:  "/",
		})
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Decode response
	var gqlResp GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&gqlResp)
	assert.NoError(t, err)

	return gqlResp
}
