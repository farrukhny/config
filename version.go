package config

import (
	"fmt"
	"strings"
	"text/tabwriter"
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

	// BuildNumber is the build number
	BuildNumber int

	// BuildTime is the time the application was built.
	BuildTime time.Time
}

// NewVersion returns a new Version struct
func NewVersion(major, minor, patch int, preRelease, commitHash, branch, metadata string, buildNumber int) Version {
	return Version{
		Major:       major,
		Minor:       minor,
		Patch:       patch,
		PreRelease:  preRelease,
		CommitHash:  commitHash,
		Branch:      branch,
		Metadata:    metadata,
		BuildNumber: buildNumber,
		BuildTime:   time.Now().UTC(),
	}
}

// Version is the current version of the application.
func (v Version) Version() string {
	// if preRelease is empty, return the version without the dash
	if v.PreRelease == "" {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%d.%d.%d-%s.%d", v.Major, v.Minor, v.Patch, v.PreRelease, v.BuildNumber)
}

// VersionInfo provides output to display the application version with more details like commit hash, branch, build time and metadata.
// it is displayed in the following format: <version> <commit hash> <branch> <build time> <metadata>
func (v Version) VersionInfo() string {
	// build version info in the table format to make it easy to read
	var versionInfo strings.Builder
	tb := tabwriter.NewWriter(&versionInfo, 0, 0, 2, ' ', tabwriter.TabIndent)
	tb.Write([]byte("Version:\t" + v.Version() + "\n"))
	if v.CommitHash != "" {
		tb.Write([]byte("Commit Hash:\t" + v.CommitHash + "\n"))
	}
	if v.Branch != "" {
		tb.Write([]byte("Branch:\t" + v.Branch + "\n"))
	}
	if !v.BuildTime.IsZero() {
		tb.Write([]byte("Build Time:\t" + v.BuildTime.Format(time.RFC3339) + "\n"))
	}
	if v.Metadata != "" {
		tb.Write([]byte(v.Metadata + "\n"))
	}
	tb.Flush()

	return versionInfo.String()
}
