package gormkratos

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/orzkratos/erkkratos"
	"github.com/orzkratos/erkkratos/erkrequire"
	"github.com/orzkratos/gormkratos/internal/errors_example"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/done"
	"github.com/yyle88/erero"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 数据库连接
var caseDB *gorm.DB

// 把错误转换为kratos错误的函数
var caseErkFc = erkkratos.NewErkFc(errors_example.ErrorServerDbTransactionError, "erk")

func TestMain(m *testing.M) {
	db := done.VCE(gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})).Nice()
	defer func() {
		done.Done(done.VCE(db.DB()).Nice().Close())
	}()

	caseDB = db
	m.Run()
}

func TestTransaction(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	}, caseErkFc)
	erkrequire.NoError(t, erk)
}

func TestTransaction_2(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorServerDbError("erx=%s", erero.New("wrong"))
	}, caseErkFc)
	erkrequire.Error(t, erk)
	//这个时候返回的错误就是函数里面的错误
	require.True(t, errors_example.IsServerDbError(erk))
}

func TestTransaction_3(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk := Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格但也符合函数本身不出错但报事务出错的情况
		return nil
	}, func(erx error) *errors.Error {
		return errors_example.ErrorServerDbTransactionError("erx=%s", erx)
	})
	erkrequire.Error(t, erk)
	//这个时候返回的错误就是事务出错
	require.True(t, errors_example.IsServerDbTransactionError(erk))
}
