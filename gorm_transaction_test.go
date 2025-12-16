// Package gormkratos_test: Tests to gormkratos transaction wrap
// Validates two-error-return pattern and transaction actions
//
// gormkratos_test: gormkratos 事务封装的测试
// 验证双错误返回模式和事务行为
package gormkratos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/errkratos/must/erkrequire"
	"github.com/orzkratos/gormkratos"
	"github.com/orzkratos/gormkratos/internal/errorspb"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/erero"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates isolated in-mem SQLite database
// setupTestDB 创建独立的内存 SQLite 数据库
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}))
	t.Cleanup(func() {
		must.Done(rese.P1(db.DB()).Close())
	})
	return db
}

// TestTransactionSuccess tests success transaction execution
// TestTransactionSuccess 测试成功的事务执行
func TestTransactionSuccess(t *testing.T) {
	db := setupTestDB(t)

	// Student represents simple test data
	// Student 表示简单的测试数据
	type Student struct {
		ID   uint   `gorm:"primarykey"`           // Main ID // 主键
		Name string `gorm:"column:name;not null"` // Student name // 学生名称
	}

	require.NoError(t, db.AutoMigrate(&Student{}))

	ctx := context.Background()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		student := &Student{Name: "test-student"}
		if err := db.Create(student).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create student: %v", err)
		}
		return nil
	})
	require.NoError(t, err)
	erkrequire.NoError(t, erk)

	// Check data was inserted
	// 检查数据已插入
	var count int64
	require.NoError(t, db.Model(&Student{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

// TestTransactionBusinessError tests business logic error handling
// TestTransactionBusinessError 测试业务逻辑错误处理
func TestTransactionBusinessError(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		return errorspb.ErrorServerDbError("business logic error: %s", erero.New("validation failed"))
	})
	require.Error(t, err)
	erkrequire.Error(t, erk)
	require.True(t, errorspb.IsServerDbError(erk))
}

// TestTransactionTimeout tests context timeout handling
// TestTransactionTimeout 测试上下文超时处理
func TestTransactionTimeout(t *testing.T) {
	db := setupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		time.Sleep(100 * time.Millisecond) // Exceed timeout // 超过超时时间
		return nil
	})
	require.Error(t, err)
	erkrequire.NoError(t, erk)
}

// TestTransactionRollback tests transaction rollback behavior
// TestTransactionRollback 测试事务回滚行为
func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)

	// Guest represents simple test data
	// Guest 表示简单的测试数据
	type Guest struct {
		ID   uint   `gorm:"primarykey"`           // Main ID // 主键
		Name string `gorm:"column:name;not null"` // Guest name // 访客名称
	}

	require.NoError(t, db.AutoMigrate(&Guest{}))

	ctx := context.Background()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		guest := &Guest{Name: "rollback-guest"}
		if err := db.Create(guest).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create guest: %v", err)
		}
		// Simulate business errors to cause rollback
		// 模拟业务错误导致回滚
		return errorspb.ErrorBadRequest("business validation failed")
	})
	require.Error(t, err)
	erkrequire.Error(t, erk)
	require.True(t, errorspb.IsBadRequest(erk))

	// Check data was rolled back, not in DB
	// 检查数据已回滚,未插入数据库
	var count int64
	require.NoError(t, db.Model(&Guest{}).Where("name = ?", "rollback-guest").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

// TestTransactionNested tests nested transaction execution
// TestTransactionNested 测试嵌套事务执行
func TestTransactionNested(t *testing.T) {
	db := setupTestDB(t)

	// Order represents test data
	// Order 表示测试数据
	type Order struct {
		ID     uint   `gorm:"primarykey"`
		Status string `gorm:"column:status;not null"`
	}

	require.NoError(t, db.AutoMigrate(&Order{}))

	ctx := context.Background()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		order := &Order{Status: "pending"}
		if err := db.Create(order).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create order: %v", err)
		}

		// Nested transaction
		// 嵌套事务
		erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
			order.Status = "confirmed"
			if err := db.Save(order).Error; err != nil {
				return errorspb.ErrorServerDbError("failed to update order: %v", err)
			}
			return nil
		})
		if err != nil {
			if erk != nil {
				return erk
			}
			return errorspb.ErrorServerDbError("nested transaction failed: %v", err)
		}

		return nil
	})
	require.NoError(t, err)
	erkrequire.NoError(t, erk)

	// Check data was inserted and updated
	// 检查数据已插入并更新
	var order Order
	require.NoError(t, db.First(&order).Error)
	require.Equal(t, "confirmed", order.Status)
}

// TestTransactionNestedRollback tests nested transaction rollback
// TestTransactionNestedRollback 测试嵌套事务回滚
func TestTransactionNestedRollback(t *testing.T) {
	db := setupTestDB(t)

	// Item represents test data
	// Item 表示测试数据
	type Item struct {
		ID    uint   `gorm:"primarykey"`
		Name  string `gorm:"column:name;not null"`
		Price int    `gorm:"column:price"`
	}

	require.NoError(t, db.AutoMigrate(&Item{}))

	ctx := context.Background()

	erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		item := &Item{Name: "phone", Price: 1000}
		if err := db.Create(item).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create item: %v", err)
		}

		// Nested transaction with rollback
		// 嵌套事务回滚
		erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
			item.Price = 900
			if err := db.Save(item).Error; err != nil {
				return errorspb.ErrorServerDbError("failed to update item: %v", err)
			}
			// Simulate nested business errors
			// 模拟嵌套业务错误
			return errorspb.ErrorBadRequest("nested validation failed")
		})
		if err != nil {
			if erk != nil {
				return erk
			}
			return errorspb.ErrorServerDbError("nested transaction failed: %v", err)
		}

		return nil
	})
	require.Error(t, err)
	erkrequire.Error(t, erk)
	require.True(t, errorspb.IsBadRequest(erk))

	// Check all data was rolled back
	// 检查所有数据已回滚
	var count int64
	require.NoError(t, db.Model(&Item{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

// TestTransactionDifferentErrors tests different error type handling
// TestTransactionDifferentErrors 测试不同错误类型处理
func TestTransactionDifferentErrors(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		db := setupTestDB(t)
		ctx := context.Background()

		erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
			return errorspb.ErrorBadRequest("invalid input")
		})
		require.Error(t, err)
		erkrequire.Error(t, erk)
		require.True(t, errorspb.IsBadRequest(erk))
	})

	t.Run("ServerDbError", func(t *testing.T) {
		db := setupTestDB(t)
		ctx := context.Background()

		erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
			return errorspb.ErrorServerDbError("database operation failed")
		})
		require.Error(t, err)
		erkrequire.Error(t, erk)
		require.True(t, errorspb.IsServerDbError(erk))
	})
}
