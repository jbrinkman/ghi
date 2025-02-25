/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package main

import "github.com/jbrinkman/ghi/cmd"

// Version information set by build using -ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
