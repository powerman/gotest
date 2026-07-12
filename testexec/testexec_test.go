package testexec_test

import (
	"testing"

	"github.com/powerman/check"

	"github.com/powerman/gotest/testexec"
)

// seen is written by helper without synchronization on purpose: if a bug in
// [testexec.Func] ever lets an unrelated top-level test run inside the same process
// as TestGuard's subtest below, both would call helper concurrently and the race
// detector will catch the resulting unsynchronized write.
var seen []string

func helper() {
	seen = append(seen, "x")
}

// TestGuard has a subtest, so its full name contains "/" - this is what used to make
// the top-level "TestGuard" component of the constructed -test.run pattern lose its
// "$" anchor, letting any other top-level test whose name has "TestGuard" as a
// literal prefix (TestGuardA, TestGuardB below) match too and get executed alongside
// it inside the very same process, racing on any state their shared code touches -
// exactly what happened with dockerize's TestFlag/TestFlagHelp/TestFlagVersion.
func TestGuard(tt *testing.T) {
	tt.Parallel()
	tt.Run("sub", func(tt *testing.T) {
		tt.Parallel()
		t := check.Must(tt)
		out, err := testexec.Func(t, helper).CombinedOutput()
		t.Nil(err)
		t.Match(out, `^$`)
	})
}

func TestGuardA(tt *testing.T) {
	tt.Parallel()
	t := check.Must(tt)
	out, err := testexec.Func(t, helper).CombinedOutput()
	t.Nil(err)
	t.Match(out, `^$`)
}

func TestGuardB(tt *testing.T) {
	tt.Parallel()
	t := check.Must(tt)
	out, err := testexec.Func(t, helper).CombinedOutput()
	t.Nil(err)
	t.Match(out, `^$`)
}

// TestGuardNested guards the same collision one level deeper: t.Run("wait") here has
// its own nested t.Run("http"), so its full name is "TestGuardNested/wait/http" - a
// naive fix that only anchors the top-level component would still leave "wait"
// unanchored, letting it match the sibling subtest "wait-list" below as a substring.
func TestGuardNested(tt *testing.T) {
	tt.Parallel()
	tt.Run("wait", func(tt *testing.T) {
		tt.Parallel()
		tt.Run("http", func(tt *testing.T) {
			tt.Parallel()
			t := check.Must(tt)
			out, err := testexec.Func(t, helper).CombinedOutput()
			t.Nil(err)
			t.Match(out, `^$`)
		})
	})
	tt.Run("wait-list", func(tt *testing.T) {
		tt.Parallel()
		t := check.Must(tt)
		out, err := testexec.Func(t, helper).CombinedOutput()
		t.Nil(err)
		t.Match(out, `^$`)
	})
}
