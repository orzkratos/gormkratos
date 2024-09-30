package gormkratos_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/orzkratos/erkkratos"
	"github.com/orzkratos/erkkratos/erkrequire"
	"github.com/orzkratos/gormkratos"
	"github.com/orzkratos/gormkratos/internal/errors_example"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/done"
	"github.com/yyle88/erero"
	"github.com/yyle88/zaplog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 数据库连接
var caseDB *gorm.DB

// 把错误转换为kratos错误的函数
var caseErkBottle = erkkratos.NewErkBottle(errors_example.ErrorServerDbTransactionError, "erk", ":->")

// 把错误转换为kratos错误的函数，这是另一种转换逻辑，把错误信息放在 metadata 里面
var caseEmtBottle = erkkratos.NewEmtBottle(errors_example.ErrorServerDbTransactionError, "wax", "erk")

func TestMain(m *testing.M) {
	db := done.VCE(gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})).Nice()
	defer func() {
		done.Done(done.VCE(db.DB()).Nice().Close())
	}()

	caseDB = db

	zaplog.LOG.Info("case")
	m.Run()
	zaplog.LOG.Info("done")
}

func TestTransaction(t *testing.T) {
	erk := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	}, caseErkBottle.Wrap)

	erkrequire.No(t, erk)
}

func TestTransaction_2(t *testing.T) {
	erk := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorServerDbError("erx=%s", erero.New("wrong"))
	}, caseErkBottle.Wrap)

	t.Log(erk)
	erkrequire.Eo(t, erk)
	//这个时候返回的错误，就是函数里面的错误
	require.True(t, errors_example.IsServerDbError(erk))
}

func TestTransaction_3(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	warpFunc := func(erx error) *errors.Error {
		return errors_example.ErrorServerDbTransactionError("erx=%s", erx)
	}

	erk := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	}, warpFunc)

	t.Log(erk)
	erkrequire.Eo(t, erk)
	//这个时候返回的错误，就是事务出错
	require.True(t, errors_example.IsServerDbTransactionError(erk))
}

func TestTransaction_4(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	}, caseErkBottle.Wrap)

	t.Log(erk)
	erkrequire.Eo(t, erk)
	//这个时候返回的错误，就是事务出错
	require.True(t, errors_example.IsServerDbTransactionError(erk))
}

func TestTransaction_5(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	}, caseEmtBottle.Wrap)

	t.Log(erk)
	erkrequire.Eo(t, erk)
	//这个时候返回的错误，就是事务出错
	require.True(t, errors_example.IsServerDbTransactionError(erk))
}

func TestTransactionV2(t *testing.T) {
	erk, err := gormkratos.TransactionV2(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	})
	require.NoError(t, err)
	erkrequire.No(t, erk) //当前面没有错误时，这里必然也没有错误
}

func TestTransactionV2_2(t *testing.T) {
	erk, err := gormkratos.TransactionV2(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorBadRequest("erx=%s", erero.New("wrong"))
	})
	t.Log(err)
	require.Error(t, err) //无论什么情况，只要有错，这里就非空，具体详见 demos 里面的样例
	t.Log(erk)
	erkrequire.Eo(t, erk)
	//这个时候返回的错误，就是函数里面的错误
	require.True(t, errors_example.IsBadRequest(erk))
}

func TestTransactionV2_3(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk, err := gormkratos.TransactionV2(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	t.Log(err)
	require.Error(t, err)
	t.Log(erk)
	erkrequire.No(t, erk) //由于是事务本身超时，这里业务上就是没有错误的
}
