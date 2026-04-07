package main

import (
	"github.com/agentop-dev/agentop/cmd"
)

var (
	version = ""
	commit  = ""
)

func main() {
	cmd.SetVersionInfo(version, commit)
	cmd.Execute()
}
