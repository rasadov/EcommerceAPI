package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dodopayments/dodopayments-go"
	"github.com/rasadov/EcommerceAPI/payment/internal"
	"github.com/rasadov/EcommerceAPI/payment/models"
	"github.com/rasadov/EcommerceAPI/payment/proto/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetCustomerByCustomerID(ctx context.Context, customerId string) (*models.Customer, error) {
	args := m.Called(ctx, customerId)
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockRepository) GetCustomerByUserID(ctx context.Context, userId uint64) (*models.Customer, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockRepository) SaveCustomer(ctx context.Context, customer *models.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockRepository) GetProductByProductID(ctx context.Context, productId string) (*models.Product, error) {
	args := m.Called(ctx, productId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockRepository) GetProductsByIDs(ctx context.Context, productIds []string) ([]*models.Product, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockRepository) SaveProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) DeleteProduct(ctx context.Context, productId string) error {
	args := m.Called(ctx, productId)
	return args.Error(0)
}

func (m *MockRepository) RegisterTransaction(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockRepository) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockRepository) Close() {}

type MockPaymentClient struct {
	mock.Mock
}

func (m *MockPaymentClient) CreateProduct(ctx context.Context, name string, price int64, currency dodopayments.Currency, taxCategory dodopayments.TaxCategory, customerId, productId string) (*dodopayments.Product, error) {
	args := m.Called(ctx, name, price, currency, taxCategory, customerId, productId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dodopayments.Product), args.Error(1)
}

func (m *MockPaymentClient) UpdateProduct(ctx context.Context, productId string, name string, price int64) error {
	args := m.Called(ctx, productId, name, price)
	return args.Error(0)
}

func (m *MockPaymentClient) ArchiveProduct(ctx context.Context, productId string) error {
	args := m.Called(ctx, productId)
	return args.Error(0)
}

func (m *MockPaymentClient) CreateCustomer(ctx context.Context, userId uint64, email, name string) (*models.Customer, error) {
	args := m.Called(ctx, userId, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockPaymentClient) CreateCustomerSession(ctx context.Context, customerId string) (string, error) {
	args := m.Called(ctx, customerId)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentClient) CreateCheckoutSession(ctx context.Context, userId uint64, customerId string, redirect string, dodoProducts []dodopayments.CheckoutSessionRequestProductCartParam, orderId uint64) (string, error) {
	args := m.Called(ctx, userId, customerId, redirect, dodoProducts, orderId)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentClient) HandleWebhook(w http.ResponseWriter, r *http.Request) (*models.Transaction, error) {
	args := m.Called(w, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func TestPaymentService_RegisterProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful product registration", func(t *testing.T) {
		dodoProduct := &dodopayments.Product{
			ProductID: "dodo-product-1",
			Price: dodopayments.Price{
				FixedPrice: 999,
				Currency:   dodopayments.CurrencyUsd,
			},
		}

		mockClient.On("CreateProduct", ctx, "Camera", int64(999), dodopayments.CurrencyUsd, dodopayments.TaxCategoryDigitalProducts, "cust-1", "product-1").
			Return(dodoProduct, nil).Once()
		mockRepo.On("SaveProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

		err := service.RegisterProduct(ctx, "Camera", 999, "cust-1", "product-1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Client error", func(t *testing.T) {
		mockClient.On("CreateProduct", ctx, "Camera", int64(999), dodopayments.CurrencyUsd, dodopayments.TaxCategoryDigitalProducts, "cust-1", "product-2").
			Return((*dodopayments.Product)(nil), errors.New("dodo error")).Once()

		err := service.RegisterProduct(ctx, "Camera", 999, "cust-1", "product-2")

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentService_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful product update", func(t *testing.T) {
		product := &models.Product{ProductID: "product-1", Price: 999}

		mockClient.On("UpdateProduct", ctx, "product-1", "Camera", int64(1299)).Return(nil).Once()
		mockRepo.On("GetProductByProductID", ctx, "product-1").Return(product, nil).Once()
		mockRepo.On("UpdateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

		err := service.UpdateProduct(ctx, "product-1", "Camera", 1299)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Client error", func(t *testing.T) {
		mockClient.On("UpdateProduct", ctx, "product-1", "Camera", int64(1299)).
			Return(errors.New("update failed")).Once()

		err := service.UpdateProduct(ctx, "product-1", "Camera", 1299)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentService_DeleteProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful product deletion", func(t *testing.T) {
		mockClient.On("ArchiveProduct", ctx, "product-1").Return(nil).Once()
		mockRepo.On("DeleteProduct", ctx, "product-1").Return(nil).Once()

		err := service.DeleteProduct(ctx, "product-1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_FindOrCreateCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Returns existing customer", func(t *testing.T) {
		customer := &models.Customer{UserId: 1, CustomerId: "cust-1"}

		mockRepo.On("GetCustomerByUserID", ctx, uint64(1)).Return(customer, nil).Once()

		result, err := service.FindOrCreateCustomer(ctx, 1, "test@example.com", "Test User")

		assert.NoError(t, err)
		assert.Equal(t, customer, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Creates new customer", func(t *testing.T) {
		customer := &models.Customer{UserId: 2, CustomerId: "cust-2", BillingEmail: "new@example.com"}

		mockRepo.On("GetCustomerByUserID", ctx, uint64(2)).Return((*models.Customer)(nil), gorm.ErrRecordNotFound).Once()
		mockClient.On("CreateCustomer", ctx, uint64(2), "new@example.com", "New User").Return(customer, nil).Once()
		mockRepo.On("SaveCustomer", ctx, customer).Return(nil).Once()

		result, err := service.FindOrCreateCustomer(ctx, 2, "new@example.com", "New User")

		assert.NoError(t, err)
		assert.Equal(t, customer, result)
		mockRepo.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentService_CreateCheckoutSession(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful checkout session", func(t *testing.T) {
		cart := []*pb.CartItem{{ProductId: "product-1", Quantity: 2}}
		products := []*models.Product{{ProductID: "product-1", DodoProductID: "dodo-1"}}

		mockRepo.On("GetProductsByIDs", ctx, []string{"product-1"}).Return(products, nil).Once()
		mockClient.On("CreateCheckoutSession", ctx, uint64(1), "cust-1", "http://localhost/redirect", mock.Anything, uint64(10)).
			Return("https://checkout.example.com", nil).Once()

		url, err := service.CreateCheckoutSession(ctx, 1, "cust-1", "http://localhost/redirect", cart, 10)

		assert.NoError(t, err)
		assert.Equal(t, "https://checkout.example.com", url)
		mockRepo.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentService_CreateCustomerPortalSession(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful portal session", func(t *testing.T) {
		customer := &models.Customer{CustomerId: "cust-1"}

		mockClient.On("CreateCustomerSession", ctx, "cust-1").Return("https://portal.example.com", nil).Once()

		url, err := service.CreateCustomerPortalSession(ctx, customer)

		assert.NoError(t, err)
		assert.Equal(t, "https://portal.example.com", url)
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentService_HandlePaymentWebhook(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockClient := new(MockPaymentClient)
	service := internal.NewPaymentService(mockClient, mockRepo)

	t.Run("Successful webhook handling", func(t *testing.T) {
		transaction := &models.Transaction{PaymentId: "pay-1", Status: models.Success.String()}
		w := httptest.NewRecorder()
		r := &http.Request{}

		mockClient.On("HandleWebhook", w, r).Return(transaction, nil).Once()
		mockRepo.On("RegisterTransaction", ctx, transaction).Return(nil).Once()

		result, err := service.HandlePaymentWebhook(ctx, w, r)

		assert.NoError(t, err)
		assert.Equal(t, transaction, result)
		mockClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
