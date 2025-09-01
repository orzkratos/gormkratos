// demo1x: Demonstration of gormkratos transaction wrapper features
// Shows various transaction scenarios including success, error, and rollback
//
// demo1x: gormkratos 事务封装功能演示
// 展示各种事务场景包括成功、错误和回滚
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/gormkratos"
	"github.com/orzkratos/gormkratos/internal/errorspb"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User represents a user model for demonstration
// User 表示演示用的用户模型
type User struct {
	ID   uint   `gorm:"primarykey"`           // Primary key // 主键
	Name string `gorm:"column:name;not null"` // User name // 用户名
	Age  int    `gorm:"column:age"`           // User age // 用户年龄
}

func main() {
	// Initialize database
	// 初始化数据库
	db := setupDatabase()
	defer rese.F0(rese.P1(db.DB()).Close)

	ctx := context.Background()

	zaplog.LOG.Info("Starting gormkratos transaction demo // 开始演示 gormkratos 事务处理")

	// Demonstrate various scenarios
	// 演示各种场景
	zaplog.LOG.Info("Executing demo // 执行演示", zap.String("scenario", "success transaction // 正常事务"))
	if erk := demoSuccessTransaction(ctx, db); erk != nil {
		zaplog.LOG.Error("Demo failed // 演示失败", zap.String("scenario", "success transaction // 正常事务"), zap.Error(erk))
	} else {
		zaplog.LOG.Info("Demo succeeded // 演示成功", zap.String("scenario", "success transaction // 正常事务"))
	}
	fmt.Println("---")

	zaplog.LOG.Info("Executing demo // 执行演示", zap.String("scenario", "business error // 业务逻辑错误"))
	if erk := demoBusinessError(ctx, db); erk != nil {
		zaplog.LOG.Error("Demo failed // 演示失败", zap.String("scenario", "business error // 业务逻辑错误"), zap.Error(erk))
	} else {
		zaplog.LOG.Info("Demo succeeded // 演示成功", zap.String("scenario", "business error // 业务逻辑错误"))
	}
	fmt.Println("---")

	zaplog.LOG.Info("Executing demo // 执行演示", zap.String("scenario", "database operations // 数据库操作"))
	if erk := demoDatabaseOperations(ctx, db); erk != nil {
		zaplog.LOG.Error("Demo failed // 演示失败", zap.String("scenario", "database operations // 数据库操作"), zap.Error(erk))
	} else {
		zaplog.LOG.Info("Demo succeeded // 演示成功", zap.String("scenario", "database operations // 数据库操作"))
	}
	fmt.Println("---")

	zaplog.LOG.Info("Executing demo // 执行演示", zap.String("scenario", "transaction rollback // 事务回滚"))
	if erk := demoTransactionRollback(ctx, db); erk != nil {
		zaplog.LOG.Error("Demo failed // 演示失败", zap.String("scenario", "transaction rollback // 事务回滚"), zap.Error(erk))
	} else {
		zaplog.LOG.Info("Demo succeeded // 演示成功", zap.String("scenario", "transaction rollback // 事务回滚"))
	}
	fmt.Println("---")

	zaplog.LOG.Info("gormkratos demo completed // gormkratos 演示完成")
}

// setupDatabase initializes in-memory SQLite database for testing
// setupDatabase 初始化用于测试的内存 SQLite 数据库
func setupDatabase() *gorm.DB {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}))

	// Auto migrate table structure
	// 自动迁移表结构
	if err := db.AutoMigrate(&User{}); err != nil {
		zaplog.LOG.Fatal("Database migration failed // 数据库迁移失败", zap.Error(err))
		os.Exit(1)
	}

	return db
}

// demoSuccessTransaction demonstrates successful transaction
// demoSuccessTransaction 演示成功的事务
func demoSuccessTransaction(ctx context.Context, db *gorm.DB) *errors.Error {
	return Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		zaplog.LOG.Info("Executing successful transaction // 执行成功的事务操作")
		return nil
	})
}

// demoBusinessError demonstrates business logic error handling
// demoBusinessError 演示业务逻辑错误处理
func demoBusinessError(ctx context.Context, db *gorm.DB) *errors.Error {
	return Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		zaplog.LOG.Info("Simulating business logic error // 模拟业务逻辑错误")
		return errorspb.ErrorBadRequest("Simulated business validation failed // 模拟的业务验证失败")
	})
}

// demoDatabaseOperations demonstrates database operations in transaction
// demoDatabaseOperations 演示事务中的数据库操作
func demoDatabaseOperations(ctx context.Context, db *gorm.DB) *errors.Error {
	return Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		user := &User{Name: "Zhang San // 张三", Age: 25}
		if err := db.Create(user).Error; err != nil {
			return errorspb.ErrorServerDbError("Failed to create user // 创建用户失败: %v", err)
		}

		zaplog.LOG.Info("Successfully created user // 成功创建用户", zap.String("name", user.Name), zap.Uint("id", user.ID))
		return nil
	})
}

// demoTransactionRollback demonstrates transaction rollback behavior
// demoTransactionRollback 演示事务回滚行为
func demoTransactionRollback(ctx context.Context, db *gorm.DB) *errors.Error {
	return Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		// Insert a record
		// 插入一条记录
		user := &User{Name: "Li Si // 李四", Age: 30}
		if err := db.Create(user).Error; err != nil {
			return errorspb.ErrorServerDbError("Failed to create user // 创建用户失败: %v", err)
		}

		zaplog.LOG.Info("Data inserted, will trigger rollback // 数据已插入，即将触发回滚", zap.String("name", user.Name))

		// Simulate business validation failure to trigger rollback
		// 模拟业务验证失败触发回滚
		return errorspb.ErrorBadRequest("Business validation failed, triggering rollback // 业务验证失败，触发事务回滚")
	})
}

// Transaction wraps gormkratos.Transaction for business layer usage
// Converts dual-error-return to single Kratos error
//
// Transaction 业务层事务封装函数
// 将双错误返回转换为单个 Kratos 错误
func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	erk, err := gormkratos.Transaction(ctx, db, run)
	if err != nil {
		if erk != nil {
			// Return business logic error
			// 返回业务逻辑错误
			return erk
		}
		// Wrap database transaction error
		// 包装数据库事务错误
		zaplog.LOG.Error("Database transaction error // 数据库事务错误", zap.Error(err))
		return errorspb.ErrorServerDbTransactionError("Transaction execution failed // 事务执行失败: %v", err)
	}
	return nil
}
