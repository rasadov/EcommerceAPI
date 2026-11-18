package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rasadov/EcommerceAPI/graphql/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func setupTestRouter(enableIntrospection bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	server := &graph.Server{}
	srv := newGraphQLHandler(server.ToExecutableSchema(), enableIntrospection)

	engine.POST("/graphql", gin.WrapH(srv))
	engine.GET("/graphql", gin.WrapH(srv))

	return engine
}

func TestIntrospectionEnabled(t *testing.T) {
	router := setupTestRouter(true)

	// Introspection query like the one GraphiQL/Playground sends to fetch docs
	query := `{"query": "{ __schema { queryType { name } } }"}`
	req, err := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(query))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp graphQLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Empty(t, resp.Errors, "should have no errors when introspection is enabled")
	assert.Contains(t, string(resp.Data), "Query", "schema introspection should return queryType")
}

func TestIntrospectionDisabled(t *testing.T) {
	router := setupTestRouter(false)

	query := `{"query": "{ __schema { queryType { name } } }"}`
	req, err := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(query))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp graphQLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.Errors, "should have errors when introspection is disabled")
	assert.Contains(t, resp.Errors[0].Message, "introspection disabled")
}

func TestIntrospectionViaGET(t *testing.T) {
	router := setupTestRouter(true)

	req, err := http.NewRequest(http.MethodGet, "/graphql?query={__schema{queryType{name}}}", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp graphQLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Empty(t, resp.Errors, "GET query should succeed when introspection is enabled")
	assert.Contains(t, string(resp.Data), "Query")
}
