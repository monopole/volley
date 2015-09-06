package net

import (
	"testing"
)

func TestFindMt(t *testing.T) {
	ns := DetermineNamespaceRoot()
	t.Logf("ns = \"%s\"", ns)
}
