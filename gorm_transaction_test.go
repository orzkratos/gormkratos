// Package gormkratos_test: Tests for gormkratos transaction wrapper
// Validates dual-error-return pattern and transaction behaviors
//
// gormkratos_test: gormkratos 事务封装的测试
// 验证双错误返回模式和事务行为
package gormkratos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/orzkratos/errkratos"
	"github.com/orzkratos/errkratos/must/erkrequire"
	"github.com/orzkratos/gormkratos"
	"github.com/orzkratos/gormkratos/internal/errorspb"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/erero"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global test database connection
// 全局测试数据库连接
var caseDB *gorm.DB

// TestMain sets up test database for all tests
// TestMain 为所有测试设置测试数据库
func TestMain(m *testing.M) {
	// Create unique in-memory SQLite database with shared cache
	// 创建具有共享缓存的唯一内存 SQLite 数据库
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}))
	defer rese.F0(rese.P1(db.DB()).Close)

	caseDB = db

	zaplog.LOG.Info("case")
	m.Run()
	zaplog.LOG.Info("done")
}

// Transaction wraps gormkratos.Transaction for business layer usage
// Converts dual-error-return to single Kratos error
//
// Transaction 为业务层使用封装 gormkratos.Transaction
// 将双错误返回转换为单个 Kratos 错误
func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errkratos.Erk) *errkratos.Erk {
	if erk, err := gormkratos.Transaction(ctx, db, run); err != nil {
		if erk != nil {
			return erk
		}
		return errorspb.ErrorServerDbTransactionError("error=%v", err)
	}
	return nil
}

// TestUser represents a simple test model
// TestUser 表示简单的测试模型
type TestUser struct {
	ID   uint   `gorm:"primarykey"`           // Primary key // 主键
	Name string `gorm:"column:name;not null"` // User name // 用户名称
}

// TestTransactionSuccess tests successful transaction execution
// TestTransactionSuccess 测试成功的事务执行
func TestTransactionSuccess(t *testing.T) {
	ctx := context.Background()
	db := caseDB

	// Auto migrate table structure
	// 自动迁移表结构
	require.NoError(t, db.AutoMigrate(&TestUser{}))

	erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
		// Insert test data
		// 插入测试数据
		user := &TestUser{Name: "test-user"}
		if err := db.Create(user).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create user: %v", err)
		}
		return nil
	})
	erkrequire.NoError(t, erk)

	// Verify data was successfully inserted
	// 验证数据已成功插入
	var count int64
	require.NoError(t, db.Model(&TestUser{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

// TestTransactionBusinessError tests business logic error handling
// TestTransactionBusinessError 测试业务逻辑错误处理
func TestTransactionBusinessError(t *testing.T) {
	ctx := context.Background()
	db := caseDB

	erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
		return errorspb.ErrorServerDbError("business logic error: %s", erero.New("validation failed"))
	})
	erkrequire.Error(t, erk)
	require.True(t, errorspb.IsServerDbError(erk))
}

// TestTransactionTimeout tests context timeout handling
// TestTransactionTimeout 测试上下文超时处理
func TestTransactionTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	db := caseDB

	erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
		time.Sleep(100 * time.Millisecond) // Exceeds timeout // 超过超时时间
		return nil
	})
	require.True(t, errorspb.IsServerDbTransactionError(erk))
}

// TestTransactionRollback tests transaction rollback behavior
// TestTransactionRollback 测试事务回滚行为
func TestTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := caseDB

	// Ensure table exists
	// 确保表存在
	require.NoError(t, db.AutoMigrate(&TestUser{}))

	// Insert data then return error to trigger rollback
	// 插入数据后返回错误触发回滚
	erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
		user := &TestUser{Name: "rollback-user"}
		if err := db.Create(user).Error; err != nil {
			return errorspb.ErrorServerDbError("failed to create user: %v", err)
		}
		// Simulate business error to trigger rollback
		// 模拟业务错误触发回滚
		return errorspb.ErrorBadRequest("business validation failed")
	})
	require.True(t, errorspb.IsBadRequest(erk))

	// Verify data was rolled back and not inserted
	// 验证数据已回滚未插入
	var count int64
	require.NoError(t, db.Model(&TestUser{}).Where("name = ?", "rollback-user").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

// TestTransactionDifferentErrors tests different error type handling
// TestTransactionDifferentErrors 测试不同错误类型处理
func TestTransactionDifferentErrors(t *testing.T) {
	ctx := context.Background()
	db := caseDB

	t.Run("BadRequest", func(t *testing.T) {
		erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
			return errorspb.ErrorBadRequest("invalid input")
		})
		erkrequire.Error(t, erk)
		require.True(t, errorspb.IsBadRequest(erk))
	})

	t.Run("ServerDbError", func(t *testing.T) {
		erk := Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
			return errorspb.ErrorServerDbError("database operation failed")
		})
		erkrequire.Error(t, erk)
		require.True(t, errorspb.IsServerDbError(erk))
	})
}
