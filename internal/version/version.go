package version

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

// Info returns version information as a map
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit_sha": CommitSHA,
		"build_time": BuildTime,
	}
}
