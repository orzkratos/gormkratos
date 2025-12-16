// Package gormkratos: GORM transaction wrap to Kratos framework
// Provides two-error-return pattern to distinguish business logic errors and database errors
//
// gormkratos: Kratos 框架的 GORM 事务封装
// 提供双错误返回模式,区分业务逻辑错误和数据库错误
package gormkratos

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/yyle88/erero"
	"gorm.io/gorm"
)

// Transaction executes a function in database transaction
// Returns two errors to distinguish transaction errors and business logic errors:
// - erk: Business logic errors (Kratos errors)
// - err: Database transaction errors
//
// IMPORTANT:
// Unlike common (res, err) pattern where res is invalid when err != nil,
// Here erk often contains valid business data when err != nil. You MUST check erk first!
//
// Error combinations:
// When err != nil:
// - erk != nil: Business logic error caused rollback
// - erk == nil: Database commit failed
// When err == nil:
// - (erk must also be nil) Both succeeded
//
// Recommended usage pattern (MUST follow this pattern):
//
//	erk, err := gormkratos.Transaction(ctx, db, run)
//	if err != nil {
//	    if erk != nil {
//	        return erk
//	    }
//	    return YourTransactionError("transaction failed: %v", err)
//	}
//
// Transaction 在数据库事务中执行函数
// 返回两个错误以区分事务错误和业务逻辑错误:
// - erk: 业务逻辑错误 (Kratos 错误)
// - err: 数据库事务错误
//
// 重要:
// 不同于常见的 (res, err) 模式中 err != nil 时 res 无效,
// 这里当 err != nil 时 erk 往往包含有效业务数据. 必须先检查 erk!
//
// 错误组合:
// 当 err != nil:
// - erk != nil: 业务逻辑错误导致回滚
// - erk == nil: 数据库提交失败
// 当 err == nil:
// - (erk 也必然是 nil) 两者都成功
//
// 推荐用法 (必须遵循此模式):
//
//	erk, err := gormkratos.Transaction(ctx, db, run)
//	if err != nil {
//	    if erk != nil {
//	        return erk
//	    }
//	    return YourTransactionError("transaction failed: %v", err)
//	}
func Transaction(
	ctx context.Context,
	db *gorm.DB,
	run func(db *gorm.DB) *errors.Error,
	options ...*sql.TxOptions,
) (erk *errors.Error, err error) {
	// Execute transaction with context and options
	// 使用上下文和选项执行事务
	if err = db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk // Business errors cause rollback // 业务错误导致回滚
		}
		return nil
	}, options...); err != nil {
		if erk != nil {
			// Business errors caused rollback, both errors back
			// 业务错误导致回滚,返回两个错误
			return erk, erero.Wro(err)
		}
		// Database errors, wrap and back
		// 数据库错误,包装后返回
		return nil, erero.Wro(err)
	}

	// Transaction succeeded
	// 事务成功
	return nil, nil
}
