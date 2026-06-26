package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/rasadov/EcommerceAPI/account/internal"
	"github.com/rasadov/EcommerceAPI/account/models"
	"github.com/rasadov/EcommerceAPI/pkg/auth"
	"github.com/rasadov/EcommerceAPI/pkg/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockRepository) PutAccount(ctx context.Context, account models.Account) (*models.Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockRepository) GetAccountByID(ctx context.Context, id uint64) (*models.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockRepository) ListAccounts(ctx context.Context, skip, take uint64) ([]*models.Account, error) {
	args := m.Called(ctx, skip, take)
	return args.Get(0).([]*models.Account), args.Error(1)
}

func (m *MockRepository) Close() {

}

func TestAccountService_Register(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	service := internal.NewService(mockRepo)

	t.Run("Successful registration", func(t *testing.T) {
		// Setup
		email := "test@example.com"
		name := "Test User"
		password := "password123"
		hashedPassword, _ := crypt.HashPassword(password)
		account := &models.Account{ID: 1, Name: name, Email: email, Password: hashedPassword}

		mockRepo.On("GetAccountByEmail", ctx, email).Return((*models.Account)(nil), errors.New("not found")).Once()
		mockRepo.On("PutAccount", ctx, mock.AnythingOfType("models.Account")).Return(account, nil).Once()

		// Execute
		result, err := service.Register(ctx, name, email, password)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		parsed, err := auth.ValidateToken(result)
		assert.NoError(t, err)
		claims, ok := parsed.Claims.(*auth.JWTCustomClaims)
		assert.True(t, ok)
		assert.Equal(t, account.ID, claims.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Account already exists", func(t *testing.T) {
		// Setup
		email := "existing@example.com"
		account := &models.Account{ID: 1, Email: email}

		mockRepo.On("GetAccountByEmail", ctx, email).Return(account, nil).Once()

		// Execute
		_, err := service.Register(ctx, "Test User", email, "password123")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "account already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestAccountService_Login(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	service := internal.NewService(mockRepo)

	t.Run("Successful login", func(t *testing.T) {
		// Setup
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := crypt.HashPassword(password)
		account := &models.Account{ID: 1, Email: email, Password: hashedPassword}

		mockRepo.On("GetAccountByEmail", ctx, email).Return(account, nil).Once()

		// Execute
		result, err := service.Login(ctx, email, password)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		parsed, err := auth.ValidateToken(result)
		assert.NoError(t, err)
		claims, ok := parsed.Claims.(*auth.JWTCustomClaims)
		assert.True(t, ok)
		assert.Equal(t, account.ID, claims.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		// Setup
		email := "test@example.com"
		password := "wrongpassword"
		hashedPassword, _ := crypt.HashPassword("correctpassword")
		account := &models.Account{ID: 1, Email: email, Password: hashedPassword}

		mockRepo.On("GetAccountByEmail", ctx, email).Return(account, nil).Once()

		// Execute
		_, err := service.Login(ctx, email, password)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAccountService_GetAccount(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	service := internal.NewService(mockRepo)

	t.Run("Successful get account", func(t *testing.T) {
		// Setup
		id := uint64(1)
		account := &models.Account{ID: id, Email: "test@example.com"}

		mockRepo.On("GetAccountByID", ctx, id).Return(account, nil).Once()

		// Execute
		result, err := service.GetAccount(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, account, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Account not found", func(t *testing.T) {
		// Setup
		id := uint64(1)
		mockRepo.On("GetAccountByID", ctx, id).Return((*models.Account)(nil), errors.New("not found")).Once()

		// Execute
		_, err := service.GetAccount(ctx, id)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAccountService_GetAccounts(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	service := internal.NewService(mockRepo)

	t.Run("Successful get accounts with valid parameters", func(t *testing.T) {
		// Setup
		skip, take := uint64(0), uint64(50)
		accounts := []*models.Account{{ID: 1}, {ID: 2}}

		mockRepo.On("ListAccounts", ctx, skip, take).Return(accounts, nil).Once()

		// Execute
		result, err := service.GetAccounts(ctx, skip, take)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, accounts, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Take exceeds limit", func(t *testing.T) {
		// Setup
		skip, take := uint64(0), uint64(150)
		accounts := []*models.Account{{ID: 1}, {ID: 2}}

		mockRepo.On("ListAccounts", ctx, skip, uint64(100)).Return(accounts, nil).Once()

		// Execute
		result, err := service.GetAccounts(ctx, skip, take)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, accounts, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Skip and take are zero", func(t *testing.T) {
		// Setup
		skip, take := uint64(0), uint64(0)
		accounts := []*models.Account{{ID: 1}, {ID: 2}}

		mockRepo.On("ListAccounts", ctx, skip, uint64(100)).Return(accounts, nil).Once()

		// Execute
		result, err := service.GetAccounts(ctx, skip, take)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, accounts, result)
		mockRepo.AssertExpectations(t)
	})
}
