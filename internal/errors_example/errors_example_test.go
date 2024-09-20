package errors_example

import (
	"testing"

	"github.com/orzkratos/erkkratos/erkrequire"
	"github.com/yyle88/erero"
)

func TestErrorServerDbError(t *testing.T) {
	erk := ErrorServerDbError("error=%v", erero.New("abc"))
	erkrequire.Eo(t, erk)
}

func TestErrorServerDbTransactionError(t *testing.T) {
	erk := ErrorServerDbTransactionError("error=%v", erero.New("abc"))
	erkrequire.Eo(t, erk)
}
