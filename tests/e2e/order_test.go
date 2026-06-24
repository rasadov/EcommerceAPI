package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 4) Create an order with 2 products
func Test04CreateOrder(t *testing.T) {
	// 1) Query products to get a list of available product IDs
	productQuery := `
        query GetProducts($pagination: PaginationInput, $query: String, $id: String, $recommended: Boolean) {
          product(pagination: $pagination, query: $query, id: $id, recommended: $recommended) {
            id
            name
            description
            price
            accountId
          }
        }
    `
	variables := map[string]interface{}{
		"pagination": map[string]interface{}{
			"skip": 0,
			"take": 5,
		},
		// Use "query" if you want to filter products by name or something, or just leave it blank
		"query":       nil,
		"id":          nil,
		"recommended": false,
	}

	productsResp := doRequest(t, serverURL, productQuery, variables)

	// 2) If there are GraphQL errors, fail immediately so we see the cause
	if len(productsResp.Errors) > 0 {
		t.Fatalf("unexpected GraphQL errors during product query: %v", productsResp.Errors)
	}

	// 3) Parse the data
	productsData, ok := productsResp.Data.(map[string]interface{})
	assert.True(t, ok, "expected product query data to be a map")

	productList, ok := productsData["product"].([]interface{})
	assert.True(t, ok, "expected 'product' field to be a slice in the response")
	assert.True(t, len(productList) >= 2, "need at least 2 products to create an order")

	// 4) Pick 2 random products
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(productList), func(i, j int) {
		productList[i], productList[j] = productList[j], productList[i]
	})
	product1 := productList[0].(map[string]interface{})
	product2 := productList[1].(map[string]interface{})

	id1, _ := product1["id"].(string)
	id2, _ := product2["id"].(string)
	assert.NotEmpty(t, id1, "product 1 id is empty")
	assert.NotEmpty(t, id2, "product 2 id is empty")

	// 5) Now, call CreateOrder using the 2 random IDs
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
					"id":       id1,
					"quantity": 2,
				},
				map[string]interface{}{
					"id":       id2,
					"quantity": 1,
				},
			},
		},
	}
	resp := doRequest(t, serverURL, createOrderQuery, orderVariables)

	// 6) Check for GraphQL errors before parsing
	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during CreateOrder")

	// 7) Assert the response is valid
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "createOrder response data should be a map")

	createdOrder, ok := data["createOrder"].(map[string]interface{})
	assert.True(t, ok, "createOrder field should be a map")

	assert.NotEmpty(t, createdOrder["id"], "expected an order ID")
	assert.NotEmpty(t, createdOrder["createdAt"], "expected a createdAt timestamp")
	assert.NotEmpty(t, createdOrder["totalPrice"], "expected a totalPrice")

	products, ok := createdOrder["products"].([]interface{})
	assert.True(t, ok, "expected products to be a list")
	assert.Len(t, products, 2, "Expected 2 products in the order")
}
