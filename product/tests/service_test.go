package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/rasadov/EcommerceAPI/product/internal"
	"github.com/rasadov/EcommerceAPI/product/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) PutProduct(ctx context.Context, p *models.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockRepository) GetProductById(ctx context.Context, id string) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockRepository) ListProducts(ctx context.Context, skip, take uint64) ([]*models.Product, error) {
	args := m.Called(ctx, skip, take)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]*models.Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockRepository) SearchProducts(ctx context.Context, query string, skip, take uint64) ([]*models.Product, error) {
	args := m.Called(ctx, query, skip, take)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockRepository) UpdateProduct(ctx context.Context, updatedProduct *models.Product) error {
	args := m.Called(ctx, updatedProduct)
	return args.Error(0)
}

func (m *MockRepository) DeleteProduct(ctx context.Context, productId string) error {
	args := m.Called(ctx, productId)
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

func TestProductService_PostProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful product creation", func(t *testing.T) {
		mockRepo.On("PutProduct", ctx, mock.AnythingOfType("*models.Product")).Run(func(args mock.Arguments) {
			product := args.Get(1).(*models.Product)
			product.ID = "product-1"
		}).Return(nil).Once()

		result, err := service.PostProduct(ctx, "Camera", "A digital camera", 99.99, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Camera", result.Name)
		assert.Equal(t, "A digital camera", result.Description)
		assert.Equal(t, 99.99, result.Price)
		assert.Equal(t, 1, result.AccountID)
		mockRepo.AssertExpectations(t)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.On("PutProduct", ctx, mock.AnythingOfType("*models.Product")).
			Return(errors.New("database error")).Once()

		result, err := service.PostProduct(ctx, "Camera", "A digital camera", 99.99, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_GetProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful get product", func(t *testing.T) {
		product := &models.Product{ID: "product-1", Name: "Camera", AccountID: 1}

		mockRepo.On("GetProductById", ctx, "product-1").Return(product, nil).Once()

		result, err := service.GetProduct(ctx, "product-1")

		assert.NoError(t, err)
		assert.Equal(t, product, result)
		mockRepo.AssertExpectations(t)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Product not found", func(t *testing.T) {
		mockRepo.On("GetProductById", ctx, "missing").Return((*models.Product)(nil), internal.ErrNotFound).Once()

		result, err := service.GetProduct(ctx, "missing")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_GetProducts(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful get products", func(t *testing.T) {
		products := []*models.Product{{ID: "1"}, {ID: "2"}}

		mockRepo.On("ListProducts", ctx, uint64(0), uint64(10)).Return(products, nil).Once()

		result, err := service.GetProducts(ctx, 0, 10)

		assert.NoError(t, err)
		assert.Equal(t, products, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_GetProductsWithIDs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful get products by IDs", func(t *testing.T) {
		ids := []string{"1", "2"}
		products := []*models.Product{{ID: "1"}, {ID: "2"}}

		mockRepo.On("ListProductsWithIDs", ctx, ids).Return(products, nil).Once()

		result, err := service.GetProductsWithIDs(ctx, ids)

		assert.NoError(t, err)
		assert.Equal(t, products, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_SearchProducts(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful search", func(t *testing.T) {
		products := []*models.Product{{ID: "1", Name: "Camera"}}

		mockRepo.On("SearchProducts", ctx, "camera", uint64(0), uint64(10)).Return(products, nil).Once()

		result, err := service.SearchProducts(ctx, "camera", 0, 10)

		assert.NoError(t, err)
		assert.Equal(t, products, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful update", func(t *testing.T) {
		existing := &models.Product{ID: "product-1", AccountID: 1}

		mockRepo.On("GetProductById", ctx, "product-1").Return(existing, nil).Once()
		mockRepo.On("UpdateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

		result, err := service.UpdateProduct(ctx, "product-1", "Updated Camera", "Updated description", 129.99, 1)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Camera", result.Name)
		assert.Equal(t, "Updated description", result.Description)
		assert.Equal(t, 129.99, result.Price)
		mockRepo.AssertExpectations(t)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Unauthorized update", func(t *testing.T) {
		existing := &models.Product{ID: "product-1", AccountID: 2}

		mockRepo.On("GetProductById", ctx, "product-1").Return(existing, nil).Once()

		result, err := service.UpdateProduct(ctx, "product-1", "Updated Camera", "Updated description", 129.99, 1)

		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	producer := newStubAsyncProducer()
	service := internal.NewProductService(mockRepo, producer)

	t.Run("Successful delete", func(t *testing.T) {
		existing := &models.Product{ID: "product-1", AccountID: 1}

		mockRepo.On("GetProductById", ctx, "product-1").Return(existing, nil).Once()
		mockRepo.On("DeleteProduct", ctx, "product-1").Return(nil).Once()

		err := service.DeleteProduct(ctx, "product-1", 1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Unauthorized delete", func(t *testing.T) {
		existing := &models.Product{ID: "product-1", AccountID: 2}

		mockRepo.On("GetProductById", ctx, "product-1").Return(existing, nil).Once()

		err := service.DeleteProduct(ctx, "product-1", 1)

		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
