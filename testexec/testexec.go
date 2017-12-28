package testexec

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

var calledFrom = map[testing.TB]bool{}

// Call let you test functions which may call os.Exit() or hang (for ex.
// main()) by running them in separate process. Result of executing f
// should be returned by process exit status and/or output to
// stdout/stderr. If f won't call os.Exit then os.Exit(0) will be called
// when f returns.
//
//	package main
//	func TestFlagHelp(tt *testing.T) {
//		t := check.T(tt)
//		out, err := testexec.Call(nil, t, main, "-h")
//		t.Match(err, "exit status 2")
//		t.Match(out, "-version")
//	}
//
// If you call Call twice per one test function it'll panic.
//
// Started process will execute same test function which calls Call up to
// this call (so this code is shared by both processes), then call f and
// exits.
//
// Each subtest started with t.Run() counts as unique test function, but
// beware subtests with t.Parallel() - they'll run after surrounding test
// function returns, so in this case both processes will share not only
// code before Call, but also code in surrounding function after t.Run()
// containing this Call.
//
// Call returns the same as (*exec.Cmd).CombinedOutput() except it convert
// stdoutStderr from []byte to string.
func Call(ctx context.Context, t testing.TB, f func(), args ...string) (stdoutStderr string, err error) {
	if calledFrom[t] {
		panic("Call can be used only once per test")
	}
	calledFrom[t] = true

	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		f()
		os.Exit(0)
	}

	args = append([]string{"-test.run=" + t.Name()}, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], args...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}
