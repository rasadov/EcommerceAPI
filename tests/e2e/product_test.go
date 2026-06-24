package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stepCreateProduct(t *testing.T) {
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
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors, "unexpected GraphQL errors during CreateProduct")

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok, "response Data should be a map")

	p, ok := data["createProduct"].(map[string]interface{})
	assert.True(t, ok, "createProduct field should be a map")

	assert.NotEmpty(t, p["id"], "expected product ID to be returned")
	assert.Equal(t, "Test Product", p["name"])
	assert.Equal(t, "A test description", p["description"])
	assert.EqualValues(t, 12.99, p["price"])
	ProductID, ok = p["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, ProductID)
	log.Println("Created product:", p)

	secondProduct := map[string]interface{}{
		"name":        "Second Product",
		"description": "Another test product",
		"price":       24.99,
	}
	resp2 := doRequest(t, query, map[string]interface{}{"product": secondProduct})
	assert.Nil(t, resp2.Errors, "unexpected GraphQL errors during second CreateProduct")

	data2, ok := resp2.Data.(map[string]interface{})
	assert.True(t, ok)
	p2, ok := data2["createProduct"].(map[string]interface{})
	assert.True(t, ok)
	ProductID2, ok = p2["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, ProductID2)
}

func stepQueryProducts(t *testing.T) {
	query := `
        query GetProducts($pagination: PaginationInput, $query: String, $id: String) {
          product(pagination: $pagination, query: $query, id: $id) {
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
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	products, ok := data["product"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(products), 2)

	log.Println("Products:", products)
}

func stepUpdateProduct(t *testing.T) {
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
			"id":          ProductID,
			"name":        "Updated Product",
			"description": "An updated description",
			"price":       15.99,
		},
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	p, ok := data["updateProduct"].(map[string]interface{})
	assert.True(t, ok)

	assert.NotEmpty(t, p["id"])
	assert.Equal(t, "Updated Product", p["name"])
	assert.Equal(t, "An updated description", p["description"])
	assert.EqualValues(t, 15.99, p["price"])
	log.Println("Updated product:", p)
}

func stepDeleteProduct(t *testing.T) {
	query := `
		mutation DeleteProduct($id: String!) {
			deleteProduct(id: $id)
		}
	`
	variables := map[string]interface{}{
		"id": ProductID,
	}

	resp := doRequest(t, query, variables)
	assert.Nil(t, resp.Errors)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	deleted, ok := data["deleteProduct"].(bool)
	assert.True(t, ok)
	assert.True(t, deleted)
	log.Println("Deleted product:", deleted)
}
