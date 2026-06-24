package main

import (
	"testing"
)

func TestE2E(t *testing.T) {
	waitForStack(t)

	t.Run("register", stepRegister)
	t.Run("login", stepLogin)
	t.Run("createProduct", stepCreateProduct)
	t.Run("createOrder", stepCreateOrder)
	t.Run("queryAccounts", stepQueryAccounts)
	t.Run("queryProducts", stepQueryProducts)
	t.Run("updateProduct", stepUpdateProduct)

	if hasDodoAPIKey() {
		t.Run("customerPortalSession", stepCustomerPortalSession)
		t.Run("checkoutSession", stepCheckoutSession)
	} else {
		t.Run("payment", func(t *testing.T) {
			t.Skip("DODO_API_KEY not set, skipping payment e2e tests")
		})
	}

	t.Run("deleteProduct", stepDeleteProduct)
}
