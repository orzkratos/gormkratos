package gormkratos

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// Transaction 该函数返回两个错误，以区分是事务错误还是使用kratos编写的业务代码逻辑有错误
func Transaction(
	ctx context.Context,
	db *gorm.DB,
	run func(db *gorm.DB) *errors.Error,
	ops ...*sql.TxOptions,
) (erk *errors.Error, err error) {
	if etx := db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk
		}
		return nil
	}, ops...); etx != nil {
		if erk != nil {
			return erk, etx
		}
		return nil, etx //因此首先判定是否有错误的，其次判定是否有kratos的错误
	}
	return nil, nil
}
