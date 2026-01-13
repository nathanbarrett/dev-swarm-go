package version

var (
	// Version is the semantic version of the build (set at compile time)
	Version = "dev"
	// Commit is the git commit hash (set at compile time)
	Commit = "unknown"
	// Date is the build date (set at compile time)
	Date = "unknown"
)

// Info returns version information as a formatted string
func Info() string {
	return Version
}

// Full returns full version information including commit and date
func Full() string {
	return Version + " (" + Commit + ") built on " + Date
}
