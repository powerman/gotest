package testinit

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey" // add goconvey support to all tests
)

// Main should be called from TestMain to ensure Setup and Teardown
// functions will be called.
//
//	func TestMain(m *testing.M) { testinit.Main(m) }
func Main(m *testing.M) {
	setup()
	code := m.Run()
	check.Report()
	teardown()
	os.Exit(code)
}

var setupFunc [8]func()

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
	if !(0 <= idx && idx < len(setupFunc)) {
		panic(fmt.Sprintf("Setup(%d) is invalid, valid values are 0…%d", idx, len(setupFunc)))
	}
	if setupFunc[idx] != nil {
		panic(fmt.Sprintf("Setup(%d) is already set", idx))
	}
	setupFunc[idx] = f
}

var teardownFunc []func()

func teardown() {
	for _, f := range teardownFunc {
		f()
	}
}

// Teardown ensure f will be called before exiting from test.
// You should always call Main from TestMain and call Fatal instead
// of log.Fatal or os.Exit.
func Teardown(f func()) {
	teardownFunc = append(teardownFunc, f)
}

// Fatal works like log.Fatal but it ensure teardown functions will be
// called before exit.
func Fatal(v ...interface{}) {
	teardown()
	log.Fatal(v...)
}
