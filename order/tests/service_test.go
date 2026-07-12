package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/rasadov/EcommerceAPI/order/internal"
	"github.com/rasadov/EcommerceAPI/order/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) PutOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) GetOrdersForAccount(ctx context.Context, accountId uint64) ([]*models.Order, error) {
	args := m.Called(ctx, accountId)
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockRepository) UpdateOrderPaymentStatus(ctx context.Context, orderId uint64, status string) error {
	args := m.Called(ctx, orderId, status)
	return args.Error(0)
}

func (m *MockRepository) Close() {}

type stubAsyncProducer struct {
	input chan *sarama.ProducerMessage
}

func newStubAsyncProducer() *stubAsyncProducer {
	p := &stubAsyncProducer{input: make(chan *sarama.ProducerMessage, 100)}
	go func() {
		for range p.input {
		}
	}()
	return p
}

func (p *stubAsyncProducer) AsyncClose() {}

func (p *stubAsyncProducer) Close() error { return nil }

func (p *stubAsyncProducer) Input() chan<- *sarama.ProducerMessage { return p.input }

func (p *stubAsyncProducer) Successes() <-chan *sarama.ProducerMessage { return nil }

func (p *stubAsyncProducer) Errors() <-chan *sarama.ProducerError { return nil }

func (p *stubAsyncProducer) IsTransactional() bool { return false }

func (p *stubAsyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag { return 0 }

func (p *stubAsyncProducer) BeginTxn() error { return nil }

func (p *stubAsyncProducer) CommitTxn() error { return nil }

func (p *stubAsyncProducer) AbortTxn() error { return nil }

func (p *stubAsyncProducer) AddOffsetsToTxn(map[string][]*sarama.PartitionOffsetMetadata, string) error {
	return nil
}

func (p *stubAsyncProducer) AddMessageToTxn(*sarama.ConsumerMessage, string, *string) error {
	return nil
}

func TestOrderService_PostOrder(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewOrderService(mockRepo, producer)

	t.Run("Successful order creation", func(t *testing.T) {
		accountID := uint64(1)
		totalPrice := 99.99
		products := []*models.OrderedProduct{
			{ID: "product-1", Quantity: 1},
			{ID: "product-2", Quantity: 2},
		}

		mockRepo.On("PutOrder", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		result, err := service.PostOrder(ctx, accountID, totalPrice, products)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, accountID, result.AccountID)
		assert.Equal(t, totalPrice, result.TotalPrice)
		assert.Equal(t, products, result.Products)
		mockRepo.AssertExpectations(t)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Repository error", func(t *testing.T) {
		accountID := uint64(2)
		totalPrice := 50.00
		products := []*models.OrderedProduct{
			{ID: "product-3", Quantity: 1},
		}

		mockRepo.On("PutOrder", ctx, mock.AnythingOfType("*models.Order")).Return(errors.New("database error")).Once()

		result, err := service.PostOrder(ctx, accountID, totalPrice, products)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetOrdersForAccount(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewOrderService(mockRepo, producer)

	t.Run("Successful get orders", func(t *testing.T) {
		accountID := uint64(1)
		orders := []*models.Order{
			{ID: 1, AccountID: accountID, TotalPrice: 99.99},
			{ID: 2, AccountID: accountID, TotalPrice: 49.50},
		}

		mockRepo.On("GetOrdersForAccount", ctx, accountID).Return(orders, nil).Once()

		result, err := service.GetOrdersForAccount(ctx, accountID)

		assert.NoError(t, err)
		assert.Equal(t, orders, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		accountID := uint64(1)

		mockRepo.On("GetOrdersForAccount", ctx, accountID).Return([]*models.Order(nil), errors.New("not found")).Once()

		result, err := service.GetOrdersForAccount(ctx, accountID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateOrderPaymentStatus(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewOrderService(mockRepo, producer)

	t.Run("Successful payment status update", func(t *testing.T) {
		orderID := uint64(1)
		status := "paid"

		mockRepo.On("UpdateOrderPaymentStatus", ctx, orderID, status).Return(nil).Once()

		err := service.UpdateOrderPaymentStatus(ctx, orderID, status)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		orderID := uint64(1)
		status := "failed"

		mockRepo.On("UpdateOrderPaymentStatus", ctx, orderID, status).Return(errors.New("update failed")).Once()

		err := service.UpdateOrderPaymentStatus(ctx, orderID, status)

		assert.Error(t, err)
		assert.Equal(t, "update failed", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
