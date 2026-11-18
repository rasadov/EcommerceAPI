package internal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/rasadov/EcommerceAPI/payment/config"
	"github.com/rasadov/EcommerceAPI/payment/dto"
	"github.com/rasadov/EcommerceAPI/payment/models"
)

type PaymentClient interface {
	CreateProduct(ctx context.Context,
		name string, price int64,
		currency dodopayments.Currency,
		taxCategory dodopayments.TaxCategory,
		customerId, productId string) (*dodopayments.Product, error)
	UpdateProduct(ctx context.Context,
		productId string,
		name string, price int64) error
	ArchiveProduct(ctx context.Context, productId string) error

	CreateCustomer(ctx context.Context, userId uint64, email, name string) (*models.Customer, error)
	CreateCustomerSession(ctx context.Context, customerId string) (string, error)

	CreateCheckoutSession(ctx context.Context,
		userId uint64,
		customerId string, redirect string,
		dodoProducts []dodopayments.CheckoutSessionRequestProductCartParam, orderId uint64) (checkoutURL string, err error)

	HandleWebhook(w http.ResponseWriter, r *http.Request) (*models.Transaction, error)
}

func NewDodoClient(apiKey string, testMode bool) PaymentClient {
	if testMode {
		return &dodoClient{
			client: dodopayments.NewClient(
				option.WithBearerToken(apiKey),
				option.WithEnvironmentTestMode(),
			),
		}
	}

	return &dodoClient{
		client: dodopayments.NewClient(
			option.WithBearerToken(apiKey),
		),
	}
}

type dodoClient struct {
	client *dodopayments.Client
}

func (d *dodoClient) CreateProduct(ctx context.Context,
	name string, price int64,
	currency dodopayments.Currency,
	taxCategory dodopayments.TaxCategory,
	customerId, productId string) (*dodopayments.Product, error) {

	product, err := d.client.Products.New(ctx, dodopayments.ProductNewParams{
		Name: dodopayments.F(name),
		Price: dodopayments.F[dodopayments.PriceUnionParam](
			dodopayments.PriceOneTimePriceParam{
				Price:    dodopayments.F(price),
				Currency: dodopayments.F(currency),
				Discount: dodopayments.F[int64](0),
			},
		),
		TaxCategory: dodopayments.F(taxCategory),
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (d *dodoClient) CreateCustomer(ctx context.Context, userId uint64, email, name string) (*models.Customer, error) {
	customer, err := d.client.Customers.New(ctx, dodopayments.CustomerNewParams{
		Email: dodopayments.F(email),
		Name:  dodopayments.F(name),
	})

	if err != nil {
		return nil, err
	}

	return &models.Customer{
		UserId:       userId,
		BillingEmail: email,
		BillingName:  name,
		CustomerId:   customer.CustomerID,
		CreatedAt:    customer.CreatedAt,
	}, nil
}

func (d *dodoClient) UpdateProduct(ctx context.Context,
	productId string,
	name string, price int64) error {

	return d.client.Products.Update(ctx, productId, dodopayments.ProductUpdateParams{
		Name: dodopayments.F(name),
		Price: dodopayments.F[dodopayments.PriceUnionParam](
			dodopayments.PriceOneTimePriceParam{
				Price:    dodopayments.F(price),
				Currency: dodopayments.F(dodopayments.CurrencyUsd),
				Discount: dodopayments.F[int64](0),
			},
		),
	})
}

func (d *dodoClient) ArchiveProduct(ctx context.Context, productId string) error {
	return d.client.Products.Archive(ctx, productId)
}

func (d *dodoClient) CreateCheckoutSession(ctx context.Context,
	userId uint64,
	customerId string, redirect string,
	dodoProducts []dodopayments.CheckoutSessionRequestProductCartParam, orderId uint64) (checkoutURL string, err error) {

	checkoutSession, err := d.client.CheckoutSessions.New(ctx, dodopayments.CheckoutSessionNewParams{
		CheckoutSessionRequest: dodopayments.CheckoutSessionRequestParam{
			Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](
				dodopayments.AttachExistingCustomerParam{
					CustomerID: dodopayments.F(customerId),
				},
			),
			ReturnURL:   dodopayments.F(redirect),
			ProductCart: dodopayments.F(dodoProducts),
			Metadata: dodopayments.F(map[string]string{
				"order_id": fmt.Sprintf("%d", orderId),
				"user_id":  fmt.Sprintf("%d", userId),
			}),
		},
	})

	if err != nil {
		return "", err
	}

	return checkoutSession.CheckoutURL, nil
}

func (d *dodoClient) CreateCustomerSession(ctx context.Context, customerId string) (string, error) {
	customerPortal, err := d.client.Customers.CustomerPortal.New(ctx, customerId,
		dodopayments.CustomerCustomerPortalNewParams{})
	if err != nil {
		return "", err
	}
	return customerPortal.Link, nil
}

func (d *dodoClient) verifyWebhookSignature(signature string, payload []byte) bool {
	h := hmac.New(sha256.New, []byte(config.DodoWebhookSecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (d *dodoClient) HandleWebhook(w http.ResponseWriter, r *http.Request) (*models.Transaction, error) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, errors.New("method not allowed")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return nil, err
	}
	defer r.Body.Close()

	webhookSignature := r.Header.Get("webhook-signature")
	if !d.verifyWebhookSignature(webhookSignature, body) {
		http.Error(w, "Invalid Webhook Signature", http.StatusBadRequest)
		return nil, errors.New("invalid Webhook Signature")
	}

	// Parse webhook payload
	var payload dto.WebhookPayload

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return nil, err
	}

	transaction := &models.Transaction{
		OrderId:      payload.Data.Metadata.OrderId,
		UserId:       payload.Data.Metadata.UserId,
		CustomerId:   payload.Data.Customer.CustomerID,
		PaymentId:    payload.Data.PaymentId,
		TotalPrice:   payload.Data.TotalAmount,
		SettledPrice: payload.Data.SettledAmount,
		Currency:     string(payload.Data.Currency),
		Status:       string(payload.Data.Status),
	}

	// Process the webhook based on event type
	switch payload.Type {
	case "payment.succeeded":
		transaction.Status = string(models.Success)
	case "payment.failed":
		transaction.Status = string(models.Failed)
	default:
		log.Printf("Unhandled webhook event type: %s", payload.Type)
	}

	// Return a 200 OK to acknowledge receipt of the webhook
	w.WriteHeader(http.StatusOK)
	return transaction, nil
}
