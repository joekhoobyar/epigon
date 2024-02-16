package test

import (
	"path"
	"runtime"
)

func FixtureDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Clean(path.Join(path.Dir(filename), "fixtures"))
}
