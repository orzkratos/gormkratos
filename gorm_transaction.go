// Package gormkratos: GORM transaction wrapper for Kratos framework
// Provides dual-error-return pattern to distinguish business logic errors from database errors
//
// gormkratos: Kratos 框架的 GORM 事务封装
// 提供双错误返回模式，区分业务逻辑错误和数据库错误
package gormkratos

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/yyle88/erero"
	"gorm.io/gorm"
)

// Transaction executes a function within a database transaction
// Returns two errors to distinguish transaction errors from business logic errors:
// - erk: Business logic error (Kratos error)
// - err: Database transaction error
//
// Transaction 在数据库事务中执行函数
// 返回两个错误以区分事务错误和业务逻辑错误：
// - erk: 业务逻辑错误 (Kratos 错误)
// - err: 数据库事务错误
func Transaction(
	ctx context.Context,
	db *gorm.DB,
	run func(db *gorm.DB) *errors.Error,
	options ...*sql.TxOptions,
) (erk *errors.Error, err error) {
	// Wrap business function to match GORM transaction signature
	// 包装业务函数以匹配 GORM 事务签名
	runFunc := func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk // Return business error to trigger rollback // 返回业务错误触发回滚
		}
		return nil
	}

	// Execute transaction with context and options
	// 使用上下文和选项执行事务
	if err = db.WithContext(ctx).Transaction(runFunc, options...); err != nil {
		if erk != nil {
			// Business error caused rollback, return both errors
			// 业务错误导致回滚，返回两个错误
			return erk, erero.Wro(err)
		}
		// Database error, return wrapped error
		// 数据库错误，返回包装后的错误
		return nil, erero.Wro(err)
	}

	// Transaction succeeded
	// 事务成功
	return nil, nil
}
