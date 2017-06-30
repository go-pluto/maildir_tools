package main

import (
	"flag"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// initLogger initializes a JSON gokit-logger set
// to the according log level supplied via CLI flag.
func initLogger(loglevel string) log.Logger {

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	switch strings.ToLower(loglevel) {
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowDebug())
	}

	return logger
}

func main() {

	// metricsPath := flag.String("metricsPath", "/metrics", "Specify where to expose collected Maildir metrics.")
	maildirRootPath := flag.String("maildirRootPath", "", "Specify path to directory containing all users' Maildirs.")
	logLevel := flag.String("logLevel", "", "Set verbosity level of logging.")
	flag.Parse()

	// Create gokit-logger based on specified verbosity level.
	logger := initLogger(*logLevel)

	if *maildirRootPath == "" {
		level.Error(logger).Log("msg", "please specify a maildirRootPath")
		os.Exit(1)
	}

	walkRootMaildir(*maildirRootPath)
}
