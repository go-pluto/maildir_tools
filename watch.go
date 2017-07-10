package main

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// watch receives fsnotify triggers on any of the
// watched directories in a user's Maildir. It
// subsequently takes action on a particular event.
func (m *UserMaildir) watch(logger log.Logger) {

	defer close(m.watchTrigger)

	for {

		select {

		case event := <-m.watcher.Events:

			if event.Op == fsnotify.Create {

				// Stat new element to check if it is
				// a directory we need to watch.
				info, err := os.Stat(event.Name)
				if err != nil {
					level.Error(logger).Log(
						"msg", "error while stat()'ing CREATE element",
						"err", err,
					)
					return
				}

				if info.IsDir() {

					// If new element is a directory,
					// add it to watcher.
					err := m.watcher.Add(event.Name)
					if err != nil {
						level.Error(logger).Log(
							"msg", "failed to add element to watcher",
							"element", event.Name,
							"err", err,
						)
					}
				}
			}

			// Trigger Maildir walk.
			m.walkTrigger <- struct{}{}

		case err := <-m.watcher.Errors:
			level.Error(logger).Log(
				"msg", "error occured while watching fsnotify triggers",
				"err", err,
			)
			return

		case <-m.done:
			level.Info(logger).Log(
				"msg", "done watching fsnotify triggers",
				"user", m.userPath,
			)
			return
		}
	}
}
