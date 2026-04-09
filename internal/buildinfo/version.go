package buildinfo

import "runtime"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Summary() string {
	return "taskmd " + Version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
}
