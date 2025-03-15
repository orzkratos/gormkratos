package main

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/orzkratos/gormkratos"
	"github.com/orzkratos/gormkratos/internal/errors_example"
	"github.com/yyle88/done"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db := done.VCE(gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})).Nice()
	defer func() {
		done.Done(done.VCE(db.DB()).Nice().Close())
	}()

	ctx := context.Background()
	if erk := runLogic(ctx, db); erk != nil {
		zaplog.LOG.Error("wrong", zap.Error(erk))
		return
	}
	zaplog.LOG.Info("success")
}

func runLogic(ctx context.Context, db *gorm.DB) *errors.Error {
	if erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		return nil
	}); erk != nil {
		zaplog.LOG.Error("wrong", zap.Error(erk))
		return erk
	}

	if erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorBadRequest("wo le ge ca")
	}); erk != nil {
		zaplog.LOG.Error("wrong", zap.Error(erk))
		return erk
	}
	return nil
}

func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	if erk, err := gormkratos.Transaction(ctx, db, run); err != nil {
		if erk != nil {
			return erk
		}
		return errors_example.ErrorServerDbTransactionError("error=%v", err)
	}
	return nil
}
