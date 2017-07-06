package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// watch receives fsnotify triggers on any of the
// watched directories in a user's Maildir. It
// subsequently takes action on a particular event.
func (m *UserMaildir) watch(logger log.Logger) {

	for {

		select {

		case event := <-m.watcher.Events:

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
					close(m.watchTrigger)
					return
				}

				if info.IsDir() {
					m.watcher.Add(event.Name)
				}

				// Trigger Maildir walk.
				m.walkTrigger <- struct{}{}

			case fsnotify.Write:

				level.Debug(logger).Log(
					"operation", "WRITE ",
					"item", event.Name,
				)

				// Trigger Maildir walk.
				m.walkTrigger <- struct{}{}

			case fsnotify.Remove:

				level.Debug(logger).Log(
					"operation", "REMOVE",
					"item", event.Name,
				)

				// Trigger Maildir walk.
				m.walkTrigger <- struct{}{}

			case fsnotify.Rename:

				level.Debug(logger).Log(
					"operation", "RENAME",
					"item", event.Name,
				)

				// Trigger Maildir walk.
				m.walkTrigger <- struct{}{}

			case fsnotify.Chmod:

				level.Debug(logger).Log(
					"operation", "CHMOD ",
					"item", event.Name,
				)

				// Trigger Maildir walk.
				m.walkTrigger <- struct{}{}
			}

		case err := <-m.watcher.Errors:
			level.Error(logger).Log(
				"msg", "error occured while watching fsnotify triggers",
				"err", err,
			)
			close(m.watchTrigger)
			return

		case <-m.done:
			level.Debug(logger).Log("msg", fmt.Sprintf("done watching fsnotify triggers for %s"))
			close(m.watchTrigger)
			return
		}
	}
}
