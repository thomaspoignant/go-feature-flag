// helper/exit_stub_test.go
package helper

// SetExitFuncForTest replaces exitFunc in tests and returns a restore func.
func SetExitFuncForTest(f func(int)) (restore func()) {
	orig := exitFunc
	exitFunc = f
	return func() { exitFunc = orig }
}
