package version

// Version information - these are set at build time
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// GetVersion returns the current version
func GetVersion() string {
	return Version
}

// GetCommit returns the current git commit hash
func GetCommit() string {
	return Commit
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return BuildDate
}
