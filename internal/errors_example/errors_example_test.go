package errors_example

import (
	"testing"

	"github.com/orzkratos/errkratos/erkrequire"
	"github.com/yyle88/erero"
)

func TestErrorServerDbError(t *testing.T) {
	erk := ErrorServerDbError("error=%v", erero.New("abc"))
	erkrequire.Error(t, erk)
}

func TestErrorServerDbTransactionError(t *testing.T) {
	erk := ErrorServerDbTransactionError("error=%v", erero.New("abc"))
	erkrequire.Error(t, erk)
}
