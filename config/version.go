package config

// version is the version of the CLI
// version is set via ldflags at build time
var version string

func Version() string {
	return version
}
