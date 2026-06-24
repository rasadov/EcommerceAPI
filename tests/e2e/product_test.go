package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// 3) Create a product
func Test03CreateProduct(t *testing.T) {
	query := `
        mutation CreateProduct($product: CreateProductInput!) {
          createProduct(product: $product) {
            id
            name
            description
            price
            accountId
          }
        }
    `
	variables := map[string]interface{}{
		"product": map[string]interface{}{
			"name":        "Test Product",
			"description": "A test description",
			"price":       12.99,
			"id":          "1",
			"accountId":   "1",
		},
	}

	resp := doRequest(t, serverURL, query, variables)
	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during CreateProduct")

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "response Data should be a map")

	p, ok := data["createProduct"].(map[string]interface{})
	assert.True(t, ok, "createProduct field should be a map")

	assert.NotEmpty(t, p["id"], "expected product ID to be returned")
	assert.Equal(t, "Test Product", p["name"])
	assert.Equal(t, "A test description", p["description"])
	assert.EqualValues(t, 12.99, p["price"])
	log.Println("Created product:", p)
}

// 4) Create an order with 2 products
func Test06QueryProducts(t *testing.T) {
	query := `
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
		// "query":       "",
		 "id":         "1",
		"recommended": false,
	}

	resp := doRequest(t, serverURL, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	products, ok := data["product"].([]interface{})
	assert.True(t, ok)

	log.Println("Products:", products)
}

func Test07UpdateProduct(t *testing.T) {
	query := `
		mutation UpdateProduct($product: UpdateProductInput!) {
			updateProduct(product: $product) {
				id
				name
				description
				price
				accountId
			}
		}
	`
	variables := map[string]interface{}{
		"product": map[string]interface{}{
			"id":          "1",
			"name":        "Updated Product",
			"description": "An updated description",
			"price":       15.99,
			"accountId":   "1",
		},
	}

	resp := doRequest(t, serverURL, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	p, ok := data["updateProduct"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, "1", p["id"])
	assert.Equal(t, "Updated Product", p["name"])
	assert.Equal(t, "An updated description", p["description"])
	assert.EqualValues(t, 15.99, p["price"])
	log.Println("Updated product:", p)
}

func Test08DeleteProduct(t *testing.T) {
	query := `
		mutation DeleteProduct($id: String!) {
			deleteProduct(id: $id) {
				id
			}
		}
	`
	variables := map[string]interface{}{
		"id": "1",
	}

	resp := doRequest(t, serverURL, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	p, ok := data["deleteProduct"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, "1", p["id"])
	log.Println("Deleted product:", p)
}

