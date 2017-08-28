package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"archive/zip"
	"io/ioutil"
	"net/http"
	"os/exec"
	"os/signal"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/oklog/pkg/group"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics aggregates all four gauges we expose
// to Prometheus for insights into underlying Maildirs.
type Metrics struct {
	duration prometheus.Histogram
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

// createMetrics initializes and registers all four
// Prometheus-exposed metrics.
func createMetrics() *Metrics {

	maildirDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "maildir_duration",
		Help:    "Duration for maildir runs",
		Buckets: []float64{.01, .02, .03, .04, .05, .06, .07, .08, .09, .10, .15, .20, .25, .30, .35, .40, .45, .50, 1},
	})

	// Register all of them with Prometheus.
	prometheus.MustRegister(maildirDuration)

	return &Metrics{
		duration: maildirDuration,
	}
}

func userDu(path string) ([]byte, error) {
	cmd := exec.Command("/usr/bin/du", "-s", path)
	return cmd.CombinedOutput()
}

// ZipFiles compresses one or many files into a single zip archive file
func ZipFiles(root string, files []os.FileInfo) (io.Reader, error) {

	newfile := bytes.NewBuffer(nil)

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		zipfile, err := os.Open(filepath.Join(root, file.Name()))
		if err != nil {
			return newfile, err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return newfile, err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return newfile, err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return newfile, err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return newfile, err
		}
	}
	return newfile, nil
}

func main() {
	// metricsPath := flag.String("metricsPath", "/metrics", "Specify where to expose collected Maildir metrics.")
	maildirRootPath := flag.String("maildirRootPath", "", "Specify path to directory containing all users' Maildirs.")
	maildirDumpPath := flag.String("maildirDumpPath", "dumps", "Specify path to directory for all 'du -s' dumps.")
	usersFlag := flag.String("users", "", "Users to watch, separated by comma.")
	intervalFlag := flag.Duration("interval", 3*time.Second, "The interval to sleep between runs.")
	workerNameFlag := flag.String("workerName", "", "The name of the worker this maildir_exporter works for.")
	logLevel := flag.String("logLevel", "", "Set verbosity level of logging.")
	flag.Parse()

	// Create gokit-logger based on specified verbosity level.
	logger := initLogger(*logLevel)

	// Create metrics struct.
	metrics := createMetrics()

	if *maildirRootPath == "" {
		level.Error(logger).Log("msg", "please specify a maildirRootPath")
		os.Exit(1)
	}

	if *usersFlag == "" {
		level.Error(logger).Log("msg", "please specify users to watch")
		os.Exit(1)
	}

	if *workerNameFlag == "" {
		level.Error(logger).Log("msg", "please specify the worker's name")
		os.Exit(1)
	}

	if err := os.MkdirAll(*maildirDumpPath, 0777); err != nil {
		level.Error(logger).Log("msg", "failed to create dump folder", "path", *maildirDumpPath, "err", err)
		os.Exit(1)
	}

	// Check that associated Google Cloud Project
	// is set as environment variable.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		level.Error(logger).Log("msg", "env flag must be set", "env", "GOOGLE_CLOUD_PROJECT")
		os.Exit(1)
	}

	// Make sure that we possess Application Default Credentials.
	appCredentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if appCredentials == "" {
		level.Error(logger).Log("msg", "env flag must be set", "env", "GOOGLE_APPLICATION_CREDENTIALS")
		os.Exit(1)
	}

	// Connect to GCS for log file uploading.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "failed to open storage client", "err", err)
		os.Exit(1)
	}

	var g group.Group
	{
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		g.Add(func() error {
			level.Debug(logger).Log("msg", "waiting for interrupt signal")
			sig := <-stop
			level.Debug(logger).Log("msg", "received sig", "signal", sig)
			return nil
		}, func(error) {})
	}
	{
		users := strings.Split(*usersFlag, ",")

		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("shit hitting the fan")
				default:
					start := time.Now()

					var combined []byte
					for _, user := range users {
						out, err := userDu(filepath.Join(*maildirRootPath, user))
						if err != nil {
							level.Warn(logger).Log(
								"msg", "failed to run 'du -s'",
								"user", user,
								"err", err,
							)
							continue
						}
						combined = append(combined, out...)
					}

					path := filepath.Join(*maildirDumpPath, fmt.Sprintf("%d", start.Unix()))
					if err := ioutil.WriteFile(path, combined, 0777); err != nil {
						level.Warn(logger).Log(
							"msg", "failed to save dump",
							"path", path,
						)
						continue
					}

					metrics.duration.Observe(time.Since(start).Seconds())
				}

				time.Sleep(*intervalFlag)
			}
		}, func(err error) {
			level.Info(logger).Log("msg", "shutting down 'du -s' loop")
			cancel()
		})
	}
	{
		// Define where we want to expose metrics via HTTP.
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{Addr: ":9275"}

		g.Add(func() error {
			level.Info(logger).Log(
				"msg", "maildir_exporter now listens for http requests",
				"addr", ":9275",
			)

			// Start HTTP server for exposing /metrics to
			// the Prometheus scraper in background.
			err := server.ListenAndServe()
			if err != nil {
				if err.Error() != "http: Server closed" {
					level.Error(logger).Log(
						"msg", "error while running HTTP server for /metrics",
						"err", err,
					)
					os.Exit(1)
				} else {
					level.Info(logger).Log("msg", "shutting down HTTP server for /metrics")
				}
			}
			return nil
		}, func(err error) {
			level.Info(logger).Log("msg", "shutting down http server")
			// Perform graceful shutdown of HTTP server.
			server.Shutdown(context.Background())
		})
	}

	if err := g.Run(); err != nil {
		level.Error(logger).Log(
			"msg", "failed to run group",
			"err", err,
		)
		os.Exit(1)
	}

	// When gracefully shutting down, upload all dumps to GCS.

	files, err := ioutil.ReadDir(*maildirDumpPath)
	if err != nil {
		level.Error(logger).Log("msg", "failed to read dump dir for uploading", "err", err)
		os.Exit(1)
	}

	zipFile, err := ZipFiles(*maildirDumpPath, files)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to create zip file",
			"err", err,
		)
		os.Exit(3)
	}

	bucket := client.Bucket("pluto-benchmark")
	obj := bucket.Object(fmt.Sprintf("maildirs/%d-%s.zip", time.Now().Unix(), *workerNameFlag)).NewWriter(ctx)
	defer func() {
		if err = obj.Close(); err != nil {
			level.Error(logger).Log("msg", "failed to close bucket object", "err", err)
		}
	}()

	_, err = io.Copy(obj, zipFile)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to upload zipfile to gcs",
			"err", err,
		)
		os.Exit(3)
	}
}
