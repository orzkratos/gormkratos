package gormkratos_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
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
	erk, err := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	})
	require.NoError(t, err)
	erkrequire.NoError(t, erk)
}

func TestTransaction_2(t *testing.T) {
	erk, err := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorServerDbError("erx=%s", erero.New("wrong"))
	})
	require.Error(t, err)
	erkrequire.Error(t, erk)
	//这个时候返回的错误，就是函数里面的错误
	require.True(t, errors_example.IsServerDbError(erk))
}

func TestTransaction_3(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk, err := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	require.Error(t, err)
	erkrequire.NoError(t, erk)
}

func TestTransaction_4(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk, err := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	require.Error(t, err)
	erkrequire.NoError(t, erk)
}

func TestTransaction_5(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk, err := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	require.Error(t, err)
	erkrequire.NoError(t, erk)
}

func TestTransaction_6(t *testing.T) {
	erk, err := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	})
	require.NoError(t, err)
	erkrequire.NoError(t, erk) //当前面没有错误时，这里必然也没有错误
}

func TestTransaction_7(t *testing.T) {
	erk, err := gormkratos.Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorBadRequest("erx=%s", erero.New("wrong"))
	})
	t.Log(err)
	require.Error(t, err) //无论什么情况，只要有错，这里就非空，具体详见 demos 里面的样例
	t.Log(erk)
	erkrequire.Error(t, erk)
	//这个时候返回的错误，就是函数里面的错误
	require.True(t, errors_example.IsBadRequest(erk))
}

func TestTransaction_8(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk, err := gormkratos.Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	t.Log(err)
	require.Error(t, err)
	t.Log(erk)
	erkrequire.NoError(t, erk) //由于是事务本身超时，这里业务上就是没有错误的
}
