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
	fc func(db *gorm.DB) *errors.Error,
	erkTxFc func(err error) *errors.Error,
	opts ...*sql.TxOptions,
) (erk *errors.Error) {
	if err := db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = fc(db); erk != nil {
			return erk
		}
		return nil
	}, opts...); err != nil {
		if erk != nil {
			return erk
		}
		return erkTxFc(err)
	}
	return nil
}
