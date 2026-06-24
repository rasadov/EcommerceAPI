package internal

import (
	"context"
	"log"

	"github.com/rasadov/EcommerceAPI/order/models"
	"gorm.io/gorm"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, order *models.Order) error
	GetOrdersForAccount(ctx context.Context, accountId uint64) ([]*models.Order, error)
	UpdateOrderPaymentStatus(ctx context.Context, orderId uint64, status string) error
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) (Repository, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Order{}, &models.ProductsInfo{})
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (repository *postgresRepository) Close() {
	sqlDB, err := repository.db.DB()
	if err == nil {
		err = sqlDB.Close()
		if err != nil {
			log.Println("Error closing postgres repository")
			log.Println(err)
		}
	}
}

func (repository *postgresRepository) PutOrder(ctx context.Context, order *models.Order) error {
	tx := repository.db.WithContext(ctx).Begin()

	err := tx.WithContext(ctx).Create(&order).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, product := range order.Products {
		orderedProduct := models.ProductsInfo{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  int(product.Quantity),
		}
		err = tx.Create(&orderedProduct).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err = tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (repository *postgresRepository) GetOrdersForAccount(ctx context.Context, accountId uint64) ([]*models.Order, error) {
	var orders []*models.Order
	err := repository.db.WithContext(ctx).
		Table("orders o").
		Select("o.id, o.created_at, o.account_id, o.total_price::money::numeric::float8, op.product_id, op.quantity").
		Joins("JOIN order_products op on o.id = op.order_id").
		Where("o.account_id = ?", accountId).
		Order("o.id").
		Scan(&orders).Error

	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (repository *postgresRepository) UpdateOrderPaymentStatus(ctx context.Context, orderId uint64, status string) error {
	return repository.db.WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", orderId).
		Update("payment_status", status).Error
}
