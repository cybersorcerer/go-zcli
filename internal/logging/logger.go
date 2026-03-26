package logger

// Package contains the log definitions to use in all zcli packages.
import (
	"log"
	"log/slog"
	"os"
)

// Global Constants to use in other packages
const (
	MyDebug   = slog.LevelDebug
	MyInfo    = slog.LevelInfo
	MyWarning = slog.LevelWarn
	MyError   = slog.LevelError
	version   = "v0.0.1b"
)

// Global variables to use in other packages
var (
	Log     *slog.Logger
	LogFile *os.File
	Lvl     *slog.LevelVar
	err     error
)

// Create logfile zcli.log, the log handler and a new instance of the handler
func init() {
	Lvl = new(slog.LevelVar)
	Lvl.Set(MyInfo)
	LogFile, err = os.OpenFile("zcli.log", os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error creating zcli log %v\n", err)
	}

	logHandler := slog.NewTextHandler(LogFile, &slog.HandlerOptions{
		Level:     Lvl,
		AddSource: true,
	})
	Log = slog.New(logHandler)
}
