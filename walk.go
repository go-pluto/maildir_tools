package main

import (
	"fmt"
	"log"
	"os"

	"io/ioutil"
	"path/filepath"
)

// MaildirItem represents the path of an element
// in the exported Maildir and its size in bytes.
type MaildirItem struct {
	Path string
	Size int64
}

// walkUserMaildir takes the Maildir root path for
// one user of the exported system as input and
// deterministically builds a file system representation
// of MaildirItems for all folders and files below
// this path.
func walkUserMaildir(userPath string) ([]MaildirItem, error) {

	items := make([]MaildirItem, 0, 10)

	err := filepath.Walk(userPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// Do not include the user's Maildir root path.
		if path == userPath {
			return nil
		}

		// Maildirs only consist of folders and files,
		// thus ignore all other elements.
		if !(info.IsDir() || info.Mode().IsRegular()) {
			return nil
		}

		// Add this file system object to the items list.
		items = append(items, MaildirItem{
			Path: path,
			Size: info.Size(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func walkRootMaildir(maildirRootPath string) {

	userMaildirs := make([][]MaildirItem, 0, 30)

	// Retrieve all file system elements in maildirRootPath.
	files, _ := ioutil.ReadDir(maildirRootPath)
	for _, f := range files {

		if f.IsDir() {

			maildirPath := filepath.Join(maildirRootPath, f.Name())

			userMaildir, err := walkUserMaildir(maildirPath)
			if err != nil {
				log.Fatal(err)
			}

			userMaildirs = append(userMaildirs, userMaildir)
		}
	}

	fmt.Println()
	for i, m := range userMaildirs {

		fmt.Printf("=== User %d ===\n", (i + 1))

		for o, e := range m {
			fmt.Printf("%2d: \"%s\" - %db\n", o, e.Path, e.Size)
		}

		fmt.Printf("\n\n")
	}
}
