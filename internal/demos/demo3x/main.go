// demo3x: Create and update operations in single transaction
// Demonstrates combining multiple database operations within one atomic transaction
//
// demo3x: 单个事务中的创建和更新操作
// 演示在一个原子事务中组合多个数据库操作
package main

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/gormkratos"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Product represents product data in the system
// Product 表示系统中的产品数据
type Product struct {
	ID    uint   `gorm:"primarykey"` // Auto-increment ID // 自增主键
	Name  string `gorm:"not null"`   // Product name, must set // 产品名称,必填
	Price int    // Product price in cents // 产品价格(分)
}

func main() {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}))
	defer rese.F0(rese.P1(db.DB()).Close)

	must.Done(db.AutoMigrate(&Product{}))

	ctx := context.Background()

	erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		product := &Product{Name: "Laptop", Price: 5000}
		if err := db.Create(product).Error; err != nil {
			return ErrorServerDbError("create failed: %v", err)
		}
		zaplog.LOG.Debug("Created product", zap.Uint("id", product.ID), zap.String("name", product.Name), zap.Int("price", product.Price))

		product.Price = 4500
		if err := db.Updates(product).Error; err != nil {
			return ErrorServerDbError("update failed: %v", err)
		}
		zaplog.LOG.Debug("Updated product", zap.Uint("id", product.ID), zap.String("name", product.Name), zap.Int("price", product.Price))
		return nil
	})
	if erk != nil {
		zaplog.LOG.Error("Error", zap.Error(erk))
	}
}

// ErrorServerDbError creates Kratos database operation errors
// ErrorServerDbError 创建 Kratos 数据库操作错误
func ErrorServerDbError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "DB_ERROR", fmt.Sprintf(format, args...))
}

// ErrorServerDbTransactionError creates Kratos transaction-level errors
// ErrorServerDbTransactionError 创建 Kratos 事务级错误
func ErrorServerDbTransactionError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "TRANSACTION_ERROR", fmt.Sprintf(format, args...))
}

// Transaction wraps gormkratos.Transaction with single-error-return pattern
// Returns business errors (erk) as-is, wraps database errors (err) as TRANSACTION_ERROR
//
// Transaction 用单错误返回模式包装 gormkratos.Transaction
// 直接返回业务错误 (erk),将数据库错误 (err) 包装为 TRANSACTION_ERROR
func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	erk, err := gormkratos.Transaction(ctx, db, run)
	if err != nil {
		if erk != nil {
			return erk
		}
		return ErrorServerDbTransactionError("transaction failed: %v", err)
	}
	return nil
}
