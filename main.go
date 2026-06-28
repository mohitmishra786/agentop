// Package main is the entry point for the agentop terminal dashboard.
package main

import (
	"github.com/agentop-dev/agentop/cmd"
)

var (
	// Version contains the application version number. It's set via ldflags when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built against. It's set via ldflags when building.
	CommitSHA = ""
)

func main() {
	cmd.SetVersionInfo(Version, CommitSHA)
	cmd.Execute()
}
