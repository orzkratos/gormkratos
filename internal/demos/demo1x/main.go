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
	db := done.VCE(gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})).Nice()
	defer func() {
		done.Done(done.VCE(db.DB()).Nice().Close())
	}()

	if erk := runLogic(db); erk != nil {
		zaplog.LOG.Info("wrong", zap.Error(erk))
	}
}

func runLogic(db *gorm.DB) *errors.Error {
	if erk, err := gormkratos.Transaction(context.Background(), db, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorBadRequest("wo le ge ca")
	}); err != nil {
		if erk != nil {
			return erk //说明是业务问题，继续报业务的错误信息
		}
		//否则说明是DB的TX提交错误
		return errors_example.ErrorServerDbTransactionError("error=%v", err)
	}
	return nil
}
