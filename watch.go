package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// Watch receives fsnotify triggers on any of the
// watched directories in a user's Maildir. It
// subsequently takes action on a particular event.
func (m *UserMaildir) Watch() {

	for {

		select {

		case event := <-m.Watcher.Events:

			switch event.Op {

			case fsnotify.Create:
				fmt.Printf("CREATE event on %v\n", event.Name)
			case fsnotify.Write:
				fmt.Printf("WRITE event on %v\n", event.Name)
			case fsnotify.Remove:
				fmt.Printf("REMOVE event on %v\n", event.Name)
			case fsnotify.Rename:
				fmt.Printf("RENAME event on %v\n", event.Name)
			case fsnotify.Chmod:
				fmt.Printf("CHMOD event on %v\n", event.Name)
			}

		case err := <-m.Watcher.Errors:
			fmt.Printf("ERROR: %v\n", err)
			return

		case <-m.done:
			fmt.Println("DONE")
			return
		}
	}
}
