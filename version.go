package config

import (
	"fmt"
	"time"
)

// Version variables are set at build time.
type Version struct {
	// Major version number
	Major int

	// Minor version number
	Minor int

	// Patch Number
	Patch int

	// PreRelease is the pre-release version
	PreRelease string

	// CommitHash git commit the build is based on
	CommitHash string

	// Branch the build is based on
	Branch string

	// Metadata is the build metadata
	Metadata string

	// BuildTime is the time the application was built.
	BuildTime time.Time
}

// NewVersion returns a new Version struct
func NewVersion(major, minor, patch int, preRelease, commitHash, branch, metadata string) Version {
	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		CommitHash: commitHash,
		Branch:     branch,
		Metadata:   metadata,
		BuildTime:  time.Now(),
	}
}

// Version is the current version of the application.
func (v Version) Version() string {
	// if preRelease is empty, return the version without the dash
	if v.PreRelease == "" {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.PreRelease)
}

// VersionInfo provides output to display the application version with more details like commit hash, branch, build time and metadata.
// it is displayed in the following format: <version> <commit hash> <branch> <build time> <metadata>
func (v Version) VersionInfo() string {
	return fmt.Sprintf("Version: v%s\nCommit Hash: %s\nBranch: %s\nBuild Time: %s\n%s", v.Version(), v.CommitHash, v.Branch, v.BuildTime, v.Metadata)
}
