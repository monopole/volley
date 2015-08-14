package table

import (
	"github.com/monopole/croupier/model"
	"testing"
)

func TestTableMake(t *testing.T) {
	expect := "3 [(3 p{0 0} v{0 0})]"
	b := NewTable(model.NewPlayer(3), nil, nil, nil, nil)
	if b.String() != expect {
		t.Errorf("Expected \"%s\", got \"%s\"", expect, b.String())
	}
}
