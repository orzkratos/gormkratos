package gormkratos

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func Transaction(
	ctx context.Context,
	db *gorm.DB,
	run func(db *gorm.DB) *errors.Error,
	efc func(error) *errors.Error,
	ops ...*sql.TxOptions,
) (erk *errors.Error) {
	if etx := db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk
		}
		return nil
	}, ops...); etx != nil {
		if erk != nil {
			return erk
		}
		return efc(etx)
	}
	return nil
}

// TransactionV2 逻辑和 Transaction 相同，该函数返回两个错误，以区分是事务错误还是使用kratos编写的业务代码逻辑有错误
// 因为 Transaction 使用 efc 函数作为参数，这样在调用时不太方便，毕竟外部调用时直接传个NewError也比较奇怪，因此想到V2
func TransactionV2(
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
