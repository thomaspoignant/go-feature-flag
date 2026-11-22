//go:build !tinygo

package flag

// cleanupsRun is whether runtime.AddCleanup runs the cleanups of collected objects.
const cleanupsRun = true
