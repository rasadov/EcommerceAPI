package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rasadov/EcommerceAPI/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

var (
	serverURL  = "http://localhost:8080/graphql"
	playground = "http://localhost:8080/playground"
	Email      string
	Password   string
	AuthToken  string
	AccountID  int
	ProductID  string
	ProductID2 string
	OrderID    int
)

func hasDodoAPIKey() bool {
	return strings.TrimSpace(os.Getenv("DODO_API_KEY")) != ""
}

func setAccountIDFromToken(t *testing.T) {
	t.Helper()
	token, err := auth.ValidateToken(AuthToken)
	require.NoError(t, err)

	claims, ok := token.Claims.(*auth.JWTCustomClaims)
	require.True(t, ok, "token claims should be JWTCustomClaims")
	AccountID = int(claims.UserID)
}

func waitForStack(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Minute)
	probeQuery := `
		query {
			accounts(pagination: { skip: 0, take: 1 }) {
				id
			}
		}
	`

	for attempt := 1; time.Now().Before(deadline); attempt++ {
		resp, err := http.Get(playground)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			t.Logf("Waiting for GraphQL gateway... (%d)", attempt)
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		gqlResp, err := postGraphQL(probeQuery, nil)
		if err == nil && len(gqlResp.Errors) == 0 {
			t.Log("Stack is ready")
			return
		}

		t.Logf("Waiting for backend services... (%d)", attempt)
		time.Sleep(10 * time.Second)
	}

	t.Fatal("stack did not become ready in time")
}

func postGraphQL(query string, variables map[string]interface{}) (GraphQLResponse, error) {
	body := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return GraphQLResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(b))
	if err != nil {
		return GraphQLResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	if AuthToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "token",
			Value: AuthToken,
			Path:  "/",
		})
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GraphQLResponse{}, err
	}
	defer resp.Body.Close()

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return GraphQLResponse{}, err
	}

	return gqlResp, nil
}

func doRequest(t *testing.T, query string, variables map[string]interface{}) GraphQLResponse {
	t.Helper()

	gqlResp, err := postGraphQL(query, variables)
	assert.NoError(t, err)
	return gqlResp
}
