package main

import (
	"testing"
)

func tlogf(t *testing.T, msg string, args ...any) {
	t.Helper()
	t.Logf("\n"+highlight+bggray+msg+resetStyle, args...)
}

func terrf(t *testing.T, msg string, args ...any) {
	t.Helper()
	t.Errorf("\n"+red+bggray+msg+resetStyle, args...)
}
