// Package testinit provides helpers to manage test setup and teardown lifecycle
// in [testing.TestMain].
package testinit

import (
	"fmt"
	"log"
	"os"
	"slices"
	"testing"

	"github.com/powerman/check"
)

// Main should be called from TestMain to ensure Setup and Teardown functions will be called.
//
//	func TestMain(m *testing.M) { testinit.Main(m) }
func Main(m *testing.M) {
	setup()
	code := m.Run()
	check.Report()
	teardown()
	os.Exit(code) //revive:disable-line:deep-exit // TestMain.
}

var setupFunc [8]func() //nolint:gochecknoglobals // Test helper requires global state.

func setup() {
	for _, f := range setupFunc {
		if f != nil {
			f()
		}
	}
}

// Setup let you run test init() functions in defined order.
//
//	func init() { testinit.Setup(1, setup) }
//	func setup() { ... }
func Setup(idx int, f func()) {
	if 0 > idx || idx >= len(setupFunc) {
		panic(fmt.Sprintf("Setup(%d) is invalid, valid values are 0…%d", idx, len(setupFunc)))
	}
	if setupFunc[idx] != nil {
		panic(fmt.Sprintf("Setup(%d) is already set", idx))
	}
	setupFunc[idx] = f
}

var teardownFunc []func() //nolint:gochecknoglobals // Test helper requires global state.

func teardown() {
	for _, v := range slices.Backward(teardownFunc) {
		v()
	}
}

// Teardown ensure f will be called before exiting from test.
// You should always call Main from TestMain and call [Fatal]
// instead of [log.Fatal] or [os.Exit].
//
// If Teardown will be called multiple times then f will be executed in reverse order -
// just like defer does.
func Teardown(f func()) {
	teardownFunc = append(teardownFunc, f)
}

// Fatal works like [log.Fatal] but it ensure teardown functions will be called before exit.
func Fatal(v ...any) {
	teardown()
	log.Fatal(v...) //revive:disable-line:deep-exit // Fatal wrapper.
}
