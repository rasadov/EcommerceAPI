package main

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func stepCustomerPortalSession(t *testing.T) {
	query := `
        mutation CreateCustomerPortalSession($credentials: CustomerPortalSessionInput!) {
          createCustomerPortalSession(credentials: $credentials) {
            url
          }
        }
    `
	variables := map[string]interface{}{
		"credentials": map[string]interface{}{
			"accountId": AccountID,
			"email":     Email,
			"name":      "John Doe",
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	session, ok := data["createCustomerPortalSession"].(map[string]interface{})
	assert.True(t, ok)

	url, ok := session["url"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, url, "expected a URL in createCustomerPortalSession response")

	log.Println("Created customer portal session:", url)
}

func stepCheckoutSession(t *testing.T) {
	query := `
		mutation CreateCheckoutSession($details: CheckoutInput!) {
			createCheckoutSession(details: $details) {
				url
			}
		}
	`
	variables := map[string]interface{}{
		"details": map[string]interface{}{
			"accountId":   AccountID,
			"email":       Email,
			"name":        "John Doe",
			"redirectUrl": "http://localhost:3000/checkout-complete",
			"orderId":     OrderID,
			"products": []map[string]interface{}{
				{"id": ProductID, "quantity": 1},
				{"id": ProductID2, "quantity": 1},
			},
		},
	}

	var resp GraphQLResponse
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp = doRequest(t, query, variables)
		if len(resp.Errors) == 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	session, ok := data["createCheckoutSession"].(map[string]interface{})
	assert.True(t, ok)

	url, ok := session["url"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, url, "expected a URL in createCheckoutSession response")

	log.Println("Created checkout session:", url)
}
