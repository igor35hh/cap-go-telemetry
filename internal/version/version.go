package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of cap-go-telemetry
	Version = "0.1.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "unknown"

	// BuildDate is the build date (set during build)
	BuildDate = "unknown"

	// GoVersion is the Go version used to build the binary
	GoVersion = runtime.Version()
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("cap-go-telemetry version %s (commit: %s, built: %s, go: %s, platform: %s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}
