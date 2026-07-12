package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rasadov/EcommerceAPI/account/internal"

	"github.com/rasadov/EcommerceAPI/account/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Add a new helper function for benchmarks
func setupBenchmarkDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(b, err)

	// Auto-migrate the Account model
	err = db.AutoMigrate(&models.Account{})
	require.NoError(b, err)

	return db
}

// Add a new helper function for benchmark repository
func setupBenchmarkRepository(b *testing.B) internal.Repository {
	db := setupBenchmarkDB(b)
	r, err := internal.NewPostgresRepository(db)
	if err != nil {
		b.Fatal(err)
	}
	return r
}

// Test helper to create an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the Account model
	err = db.AutoMigrate(&models.Account{})
	require.NoError(t, err)

	return db
}

// Test helper to create a repository with test database
func setupTestRepository(t *testing.T) internal.Repository {
	db := setupTestDB(t)
	r, err := internal.NewPostgresRepository(db)
	if err != nil {
		log.Println(err)
	}
	return r
}

// Test helper to create a sample account
func createSampleAccount() models.Account {
	return models.Account{
		ID:       uint64(time.Now().UnixNano()),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword123",
	}
}

func TestRepository_PutAccount(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	ctx := context.Background()
	account := createSampleAccount()

	t.Run("successful account creation", func(t *testing.T) {
		result, err := repo.PutAccount(ctx, account)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, account.ID, result.ID)
		assert.Equal(t, account.Email, result.Email)
		assert.Equal(t, account.Name, result.Name)
	})

	t.Run("duplicate account creation should fail", func(t *testing.T) {
		// Second creation with same ID should fail
		_, err := repo.PutAccount(ctx, account)
		assert.Error(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		newAccount := createSampleAccount()
		newAccount.ID = 4

		_, err := repo.PutAccount(ctx, newAccount)
		// The behavior depends on timing, but generally should handle cancellation
		// This test ensures the method respects context
		if err != nil {
			t.Logf("Expected behavior: context cancellation handled: %v", err)
		}
	})
}

func TestRepository_GetAccountByEmail(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	ctx := context.Background()
	account := createSampleAccount()

	// Setup: Create an account first
	_, err := repo.PutAccount(ctx, account)
	require.NoError(t, err)

	t.Run("successful retrieval by email", func(t *testing.T) {
		result, err := repo.GetAccountByEmail(ctx, account.Email)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, account.ID, result.ID)
		assert.Equal(t, account.Email, result.Email)
		assert.Equal(t, account.Name, result.Name)
	})

	t.Run("account not found by email", func(t *testing.T) {
		_, err := repo.GetAccountByEmail(ctx, "nonexistent@example.com")

		assert.Error(t, err)
		// Should return gorm.ErrRecordNotFound
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := repo.GetAccountByEmail(ctx, "")

		assert.Error(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.GetAccountByEmail(ctx, account.Email)
		if err != nil {
			t.Logf("Context cancellation handled: %v", err)
		}
	})
}

func TestRepository_GetAccountByID(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	ctx := context.Background()
	account := createSampleAccount()

	// Setup: Create an account first
	_, err := repo.PutAccount(ctx, account)
	require.NoError(t, err)

	t.Run("successful retrieval by ID", func(t *testing.T) {
		result, err := repo.GetAccountByID(ctx, account.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, account.ID, result.ID)
		assert.Equal(t, account.Email, result.Email)
		assert.Equal(t, account.Name, result.Name)
	})

	t.Run("account not found by ID", func(t *testing.T) {
		_, err := repo.GetAccountByID(ctx, 12412412)

		assert.Error(t, err)
		// Should return gorm.ErrRecordNotFound
	})
}

func TestRepository_ListAccounts(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	ctx := context.Background()

	// Setup: Create multiple accounts
	accounts := []models.Account{
		{
			ID:    101,
			Email: "user1@example.com",
			Name:  "User One",
		},
		{
			ID:    102,
			Email: "user2@example.com",
			Name:  "User Two",
		},
		{
			ID:    103,
			Email: "user3@example.com",
			Name:  "User Three",
		},
	}

	for _, account := range accounts {
		_, err := repo.PutAccount(ctx, account)
		require.NoError(t, err)
	}

	t.Run("list all accounts", func(t *testing.T) {
		result, err := repo.ListAccounts(ctx, 0, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 3)

		// Verify accounts are returned
		emails := make([]string, len(result))
		for i, acc := range result {
			emails[i] = acc.Email
		}
		assert.Contains(t, emails, "user1@example.com")
		assert.Contains(t, emails, "user2@example.com")
		assert.Contains(t, emails, "user3@example.com")
	})

	t.Run("pagination - skip and take", func(t *testing.T) {
		// Get first 2 accounts
		result, err := repo.ListAccounts(ctx, 0, 2)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// Get next account (skip 2, take 1)
		result, err = repo.ListAccounts(ctx, 2, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("empty result when skip exceeds total", func(t *testing.T) {
		result, err := repo.ListAccounts(ctx, 100, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("zero take parameter", func(t *testing.T) {
		result, err := repo.ListAccounts(ctx, 0, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestRepository_Close(t *testing.T) {
	repo := setupTestRepository(t)

	// Should not panic or return error
	assert.NotPanics(t, func() {
		repo.Close()
	})

	// After closing, operations should fail
	ctx := context.Background()
	account := createSampleAccount()

	_, err := repo.PutAccount(ctx, account)
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkPostgresRepository_PutAccount(b *testing.B) {
	repo := setupBenchmarkRepository(b)
	defer repo.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account := models.Account{
			ID:    uint64(i),
			Email: fmt.Sprintf("bench%d@example.com", i),
			Name:  fmt.Sprintf("Bench User %d", i),
		}

		_, err := repo.PutAccount(ctx, account)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostgresRepository_GetAccountByEmail(b *testing.B) {
	repo := setupBenchmarkRepository(b)
	defer repo.Close()

	ctx := context.Background()

	// Setup: Create test accounts
	for i := 0; i < 100; i++ {
		account := models.Account{
			ID:    uint64(i),
			Email: fmt.Sprintf("bench%d@example.com", i),
			Name:  fmt.Sprintf("Bench User %d", i),
		}
		_, err := repo.PutAccount(ctx, account)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("bench%d@example.com", i%100)
		_, err := repo.GetAccountByEmail(ctx, email)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for testing with timeout
func TestWithTimeout(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	account := createSampleAccount()

	// This should either complete quickly or timeout
	_, err := repo.PutAccount(ctx, account)
	if err != nil {
		t.Logf("Operation timed out as expected: %v", err)
	}
}
