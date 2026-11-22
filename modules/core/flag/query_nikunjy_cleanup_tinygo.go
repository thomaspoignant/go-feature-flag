//go:build tinygo

package flag

// cleanupsRun is whether runtime.AddCleanup runs the cleanups of collected objects. TinyGo
// accepts the call and never runs the cleanup (checked with TinyGo 0.40.1), so keys that
// are released by a cleanup would stay for the life of the process.
const cleanupsRun = false
