package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"os/signal"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// UserMaildir collects the exported metrics and
// the internal representation of the underlying
// Maildir structure for one user in the system.
type UserMaildir struct {
	done     chan struct{}
	Metrics  map[string]uint64
	Checksum []byte
	Items    []MaildirItem
	Watcher  *fsnotify.Watcher
}

// MaildirItem represents the path of an element
// in the exported Maildir and its size in bytes.
type MaildirItem struct {
	Path string
	Size uint64
}

// initLogger initializes a JSON gokit-logger set
// to the according log level supplied via CLI flag.
func initLogger(loglevel string) log.Logger {

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.Caller(5),
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

	// Create a signal channel to handle program termination.
	sigs := make(chan os.Signal, 1)

	// Register sigs channel to be triggered
	// when SIGINT or SIGTERM are invoked.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Retrieve internal representation of all
	// folders and files per user in specified
	// Maildir directory.
	userMaildirs, err := walkRootMaildir(*maildirRootPath)
	if err != nil {
		level.Error(logger).Log(
			"msg", fmt.Sprintf("failed to walk user Maildirs at %s", *maildirRootPath),
			"err", err,
		)
	}

	// Kick-off fsnotify trigger processing for
	// all watched Maildirs.
	for _, m := range *userMaildirs {
		go Watch(logger, m.Watcher, m.done)
	}

	// Wait until we receive a program termination.
	<-sigs
	fmt.Println()

	// Instruct watcher to finish.
	for _, m := range *userMaildirs {
		m.done <- struct{}{}
	}
}
