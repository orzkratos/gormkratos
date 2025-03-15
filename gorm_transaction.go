package gormkratos

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/yyle88/erero"
	"gorm.io/gorm"
)

// Transaction 该函数返回两个错误，以区分是事务错误还是使用kratos编写的业务代码逻辑有错误
func Transaction(
	ctx context.Context,
	db *gorm.DB,
	run func(db *gorm.DB) *errors.Error,
	options ...*sql.TxOptions,
) (erk *errors.Error, err error) {
	runFunc := func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk
		}
		return nil
	}
	if err = db.WithContext(ctx).Transaction(runFunc, options...); err != nil {
		if erk != nil {
			return erk, erero.Wro(err)
		}
		return nil, erero.Wro(err) //因此首先判定是否有错误的，其次判定是否有kratos的错误
	}
	return nil, nil
}
