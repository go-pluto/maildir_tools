package main

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Watch receives fsnotify triggers on any of the
// watched directories in a user's Maildir. It
// subsequently takes action on a particular event.
func Watch(logger log.Logger, w *fsnotify.Watcher, done chan struct{}) {

	for {

		select {

		case event := <-w.Events:

			switch event.Op {

			case fsnotify.Create:
				level.Debug(logger).Log(
					"operation", "CREATE",
					"item", event.Name,
				)

				// Stat new element to check if it is
				// a directory we need to watch.
				info, err := os.Stat(event.Name)
				if err != nil {
					level.Error(logger).Log(
						"msg", "error while stat()'ing CREATE element",
						"err", err,
					)
				}

				if info.IsDir() {
					w.Add(event.Name)
				}

			case fsnotify.Write:
				level.Debug(logger).Log(
					"operation", "WRITE ",
					"item", event.Name,
				)

			case fsnotify.Remove:
				level.Debug(logger).Log(
					"operation", "REMOVE",
					"item", event.Name,
				)

			case fsnotify.Rename:
				level.Debug(logger).Log(
					"operation", "RENAME",
					"item", event.Name,
				)

			case fsnotify.Chmod:
				level.Debug(logger).Log(
					"operation", "CHMOD ",
					"item", event.Name,
				)
			}

		case err := <-w.Errors:
			level.Error(logger).Log(
				"msg", "error occured while watching fsnotify triggers",
				"err", err,
			)
			return

		case <-done:
			level.Debug(logger).Log("msg", "done watching fsnotify triggers")
			return
		}
	}
}
