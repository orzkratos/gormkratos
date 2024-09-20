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
	if err := db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if erk = run(db); erk != nil {
			return erk
		}
		return nil
	}, ops...); err != nil {
		if erk != nil {
			return erk
		}
		return efc(err)
	}
	return nil
}
