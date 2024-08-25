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
	newErkTx func(error) *errors.Error,
	opts ...*sql.TxOptions,
) (erk *errors.Error) {
	if etx := db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = fc(db); erk != nil {
			return erk
		}
		return nil
	}, opts...); etx != nil {
		if erk != nil {
			return erk
		}
		return newErkTx(etx)
	}
	return nil
}
