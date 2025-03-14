package gormkratos_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/orzkratos/errkratos/erkrequire"
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
	db := done.VCE(gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{
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

// 在你的业务层再封装一次把你的"数据库事务报错"的错误像这样包装进来就行
func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	if erk, err := gormkratos.Transaction(ctx, db, run); err != nil {
		if erk != nil {
			return erk
		}
		return errors_example.ErrorServerDbTransactionError("error=%v", err)
	}
	return nil
}

func TestTransaction(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	})
	erkrequire.NoError(t, erk)
}

func TestTransaction_2(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorServerDbError("erx=%s", erero.New("wrong"))
	})
	erkrequire.Error(t, erk)
	//这个时候返回的错误，就是函数里面的错误
	require.True(t, errors_example.IsServerDbError(erk))
}

func TestTransaction_3(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancelFunc()

	erk := Transaction(ctx, caseDB, func(db *gorm.DB) *errors.Error {
		time.Sleep(time.Millisecond * 100)
		//其实这种不太严格，但也符合函数本身不出错，但报事务出错的情况
		return nil
	})
	require.True(t, errors_example.IsServerDbTransactionError(erk))
}

func TestTransaction_4(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return nil
	})
	erkrequire.NoError(t, erk) //当前面没有错误时，这里必然也没有错误
}

func TestTransaction_5(t *testing.T) {
	erk := Transaction(context.Background(), caseDB, func(db *gorm.DB) *errors.Error {
		return errors_example.ErrorBadRequest("erx=%s", erero.New("wrong"))
	})
	require.True(t, errors_example.IsBadRequest(erk))
}
