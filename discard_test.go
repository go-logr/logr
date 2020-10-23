package logr

import (
	"errors"
	"testing"
)

func TestDiscard(t *testing.T) {
	l := Discard()
	if _, ok := l.(discardLogger); !ok {
		t.Error("did not return the expected underlying type")
	}
	// Verify that none of the methods panic, there is not more we can test.
	l.WithName("discard").WithValues("z", 5).Info("Hello world")
	l.Info("Hello world", "x", 1, "y", 2)
	l.V(1).Error(errors.New("foo"), "a", 123)
	if l.Enabled() {
		t.Error("discard logger must always say it is disabled")
	}
}
