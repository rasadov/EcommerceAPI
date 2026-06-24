package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func stepCreateOrder(t *testing.T) {
	assert.NotEmpty(t, ProductID, "ProductID must be set before creating an order")
	assert.NotEmpty(t, ProductID2, "ProductID2 must be set before creating an order")

	createOrderQuery := `
        mutation CreateOrder($order: OrderInput!) {
          createOrder(order: $order) {
            id
            createdAt
            totalPrice
            products {
                id
                name
                price
                quantity
            }
          }
        }
    `
	orderVariables := map[string]interface{}{
		"order": map[string]interface{}{
			"products": []interface{}{
				map[string]interface{}{
					"id":       ProductID,
					"quantity": 2,
				},
				map[string]interface{}{
					"id":       ProductID2,
					"quantity": 1,
				},
			},
		},
	}

	var resp GraphQLResponse
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp = doRequest(t, createOrderQuery, orderVariables)
		if len(resp.Errors) == 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during CreateOrder")

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "createOrder response data should be a map")

	createdOrder, ok := data["createOrder"].(map[string]interface{})
	assert.True(t, ok, "createOrder field should be a map")

	assert.NotEmpty(t, createdOrder["id"], "expected an order ID")
	assert.NotEmpty(t, createdOrder["createdAt"], "expected a createdAt timestamp")
	assert.NotEmpty(t, createdOrder["totalPrice"], "expected a totalPrice")

	orderID, ok := createdOrder["id"].(float64)
	assert.True(t, ok)
	OrderID = int(orderID)

	products, ok := createdOrder["products"].([]interface{})
	assert.True(t, ok, "expected products to be a list")
	assert.Len(t, products, 2, "Expected 2 products in the order")
}
