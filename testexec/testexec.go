// Package testexec provides helpers to test functions which may call [os.Exit]
// or hang by running them in a separate process.
package testexec

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
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
//		out, err := testexec.Func(t, main, "-h").CombinedOutput()
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
func Func(tb testing.TB, f func(), args ...string) *exec.Cmd {
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

	args = append([]string{"-test.run=" + runPattern(tb.Name())}, args...)
	cmd := exec.CommandContext(tb.Context(), os.Args[0], args...) //nolint:gosec // Re-executing test binary is by design.
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

// runPattern builds a -test.run value which matches only the test identified by name
// (as returned by [testing.TB.Name]), not any other test whose name
// happens to share a "/"-delimited component prefix with it, at any depth.
//
// The -test.run flag splits its value on "/" into one regexp per element
// and matches each in isolation, without requiring a match to span the whole element.
// Go computes the elements to match against the very same way,
// by splitting the test's own full name on every "/"
// (see the elem variable in matcher.fullName in testing/match.go) -
// including any "/" that's merely part of a subtest's own descriptive name
// rather than a t.Run boundary.
// So a naive "^"+name+"$" only anchors the first and last of these elements,
// leaving every other element - and, whenever name identifies a subtest,
// the first one too - unanchored and matchable as a mere substring,
// which lets unrelated tests or subtests whose name shares such a substring match too
// and run alongside the intended one, sharing this process' [os.Args] and racing on it.
//
// Anchoring every element individually closes this for good:
// since Go derives its own elements the same way,
// this reproduces its match exactly for the intended test
// while rejecting any other whose elements diverge at any position.
func runPattern(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		parts[i] = "^" + regexp.QuoteMeta(p) + "$"
	}
	return strings.Join(parts, "/")
}
