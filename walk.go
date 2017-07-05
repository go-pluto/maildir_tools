package main

import (
	"os"

	"crypto/sha512"
	"io/ioutil"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// walkUserMaildir takes the Maildir root path for
// one user of the exported system as input and
// deterministically builds a file system representation
// of MaildirItems for all folders and files below
// this path. It also calculates initial metrics.
func (m *UserMaildir) walkUserMaildir(userPath string) error {

	shaHash := sha512.New()
	m.Items = make([]MaildirItem, 0, 10)

	err := filepath.Walk(userPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// Do not include the user's Maildir root path,
		// but add it to this user's file system watcher.
		if path == userPath {

			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			err = m.Watcher.Add(absPath)
			if err != nil {
				return err
			}

			return nil
		}

		// Maildirs only consist of folders and files,
		// thus ignore all other elements.
		if !(info.IsDir() || info.Mode().IsRegular()) {
			return nil
		}

		size := uint64(info.Size())

		// Add this file system object to the items list.
		m.Items = append(m.Items, MaildirItem{
			Path: path,
			Size: size,
		})

		if info.IsDir() {

			m.Metrics["maildir_folders_total"]++

			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			// Add this sub directory to this user's watcher.
			err = m.Watcher.Add(absPath)
			if err != nil {
				return err
			}
		} else if info.Mode().IsRegular() {
			m.Metrics["maildir_files_total"]++
		}

		m.Metrics["maildir_elements_total"]++
		m.Metrics["maildir_size_total"] += size

		shaHash.Write([]byte(path))

		return nil
	})
	if err != nil {
		return err
	}

	m.Checksum = shaHash.Sum(nil)

	return nil
}

// walkRootMaildir expects the path to the monitored
// node's Maildir directory containing one folder per
// user in the system. It then walks each user folder
// individually and deterministically by invoking the
// above function walkUserMaildir.
func walkRootMaildir(maildirRootPath string) (*[]UserMaildir, error) {

	i := 0
	userMaildirs := make([]UserMaildir, 0, 30)

	// Retrieve all file system elements in maildirRootPath.
	files, _ := ioutil.ReadDir(maildirRootPath)
	for _, f := range files {

		if f.IsDir() {

			// Create new file system watcher for this user.
			w, err := fsnotify.NewWatcher()
			if err != nil {
				return nil, err
			}

			// Create new item for this user.
			userMaildirs = append(userMaildirs, UserMaildir{
				done:    make(chan struct{}),
				Metrics: make(map[string]uint64),
				Items:   nil,
				Watcher: w,
			})

			// Set metrics initially to zero.
			userMaildirs[i].Metrics["maildir_elements_total"] = 0
			userMaildirs[i].Metrics["maildir_folders_total"] = 0
			userMaildirs[i].Metrics["maildir_files_total"] = 0
			userMaildirs[i].Metrics["maildir_size_total"] = 0

			maildirPath := filepath.Join(maildirRootPath, f.Name())

			// Inspect user folder content individually.
			err = userMaildirs[i].walkUserMaildir(maildirPath)
			if err != nil {
				return nil, err
			}

			i++
		}
	}

	return &userMaildirs, nil
}
