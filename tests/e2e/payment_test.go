package main

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCustomerPortalSession(t *testing.T) {
	query := `
        mutation CreateCustomerPortalSession($accountId: String!) {
          createCustomerPortalSession(accountId: $accountId) {
            url
          }
        }
    `
	Email = fmt.Sprintf("random%d@example.com", rand.Intn(100000))
	variables := map[string]interface{}{
		"accountId": "1",
		"email":     Email,
		"name":      "John Doe",
	}

	resp := doRequest(t, serverURL, query, variables)
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

func TestCheckoutSession(t *testing.T) {
	query := `
		mutation CreateCheckoutSession($accountId: String!, $products: [CheckoutProductInput!]!) {
			createCheckoutSession(accountId: $accountId, products: $products, email: $email, name: $name, redirectUrl: $redirectUrl, orderId: $orderId) {
				url
			}
		}
	`
	variables := map[string]interface{}{
		"accountId":   "1",
		"email":       Email,
		"name":        "John Doe",
		"redirectUrl": "http://localhost:3000/checkout-complete",
		"orderId":     1,
		"products": []map[string]interface{}{
			{"id": "1", "quantity": 1},
			{"id": "2", "quantity": 1},
		},
	}

	resp := doRequest(t, serverURL, query, variables)
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
