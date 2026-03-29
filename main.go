package main

import (
	"zcli/cmd"
	logger "zcli/internal/logging"
)

var (
	version = "v0.5.0"
	commit  = "unknown"
)

func main() {
	defer logger.LogFile.Close()
	logger.Log.Info("zCLI Version", "version", version, "commit", commit)
	cmd.SetVersionInfo(version, commit)
	cmd.Execute()
}
