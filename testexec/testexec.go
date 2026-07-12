// Package testexec provides helpers to test functions which may call [os.Exit]
// or hang by running them in a separate process.
package testexec

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"testing"
)

//nolint:gochecknoglobals // Test helper requires global state to track per-test usage.
var (
	calledFrom   = make(map[testing.TB]bool)
	calledFromMu sync.Mutex
)

// Func let you test functions which may call [os.Exit]() or hang (for ex. main())
// by running them in separate process.
// Result of executing f should be returned by process exit status
// and/or output to stdout/stderr.
// If f won't call [os.Exit] then [os.Exit](0) will be called when f returns.
//
//	package main
//	func TestFlagHelp(tt *testing.T) {
//		t := check.Must(tt)
//		out, err := testexec.Func(ctx, t, main, "-h").CombinedOutput()
//		t.Match(err, "exit status 2")
//		t.Match(out, "-version")
//	}
//
// If you call Func twice per one test function it'll panic.
//
// Started process will execute same test function which calls Func up to this call
// (so this code is shared by both processes), then call f and exits.
//
// Each subtest started with t.Run() counts as unique test function,
// but beware subtests with t.Parallel() - they'll run after surrounding test function returns,
// so in this case both processes will share not only code before Func,
// but also code in surrounding function after t.Run() containing this Func.
func Func(ctx context.Context, tb testing.TB, f func(), args ...string) *exec.Cmd {
	tb.Helper()

	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		f()
		os.Exit(0) //revive:disable-line:deep-exit // This runs in a subprocess.
	}

	calledFromMu.Lock()
	defer calledFromMu.Unlock()
	if calledFrom[tb] {
		panic("Func can be used only once per test")
	}
	calledFrom[tb] = true

	args = append([]string{"-test.run=^" + regexp.QuoteMeta(tb.Name()) + "$"}, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], args...) //nolint:gosec // Re-executing test binary is by design.
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}
