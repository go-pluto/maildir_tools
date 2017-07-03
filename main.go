package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// UserMaildir collects the exported metrics and
// the internal representation of the underlying
// Maildir structure for one user in the system.
type UserMaildir struct {
	Metrics  map[string]uint64
	Checksum []byte
	Items    []MaildirItem
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

	fmt.Println()
	for i, m := range *userMaildirs {

		fmt.Printf("=== User %d ===\n\n", (i + 1))

		for k, p := range m.Metrics {
			fmt.Printf("%s => %d\n", k, p)
		}

		fmt.Printf("sha512 checksum => %x\n\n", m.Checksum)

		for o, e := range m.Items {
			fmt.Printf("%2d: \"%s\" - %db\n", o, e.Path, e.Size)
		}

		fmt.Printf("\n\n\n")
	}
}
