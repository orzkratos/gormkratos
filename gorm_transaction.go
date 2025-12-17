// Package gormkratos: GORM transaction integration with Kratos
// Provides two-error-return pattern to distinguish business logic errors and database errors
//
// gormkratos: GORM 事务与 Kratos 集成
// 提供双错误返回模式, 区分业务逻辑错误和数据库错误
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
// When err != nil and erk != nil, erk contains the specific business reason.
// Return erk first since it has more business context (reason and code) than what the raw transaction throws.
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
// 当 err != nil 且 erk != nil 时, erk 包含业务层的具体原因.
// 需要优先返回 erk, 因为它比底层事务抛出的错误更有业务错误原因和错误码信息.
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
			// Business error caused rollback, return both errors
			// 业务错误导致回滚, 返回两个错误
			return erk, erero.Wro(err)
		}
		// Database error, wrap and return
		// 数据库错误, 包装后返回
		return nil, erero.Wro(err)
	}

	// Transaction succeeded
	// 事务成功
	return nil, nil
}
