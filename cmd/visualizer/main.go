package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("You must provide two files to compare")
	}

	files := os.Args[1:3]

	data := make(map[string]map[string]int)

	for _, file := range files {
		if err := readZip(file, data); err != nil {
			log.Fatal(err)
		}
	}

	header := "import matplotlib.pyplot as plot\n\n"
	buf := bytes.NewBufferString(header)

	if err := matplotlibWriter(buf, data); err != nil {
		log.Fatal(err)
	}

	//if err := matplotlibLegendWriter(buf, results); err != nil {
	//	return err
	//}

	footer := "plot.grid(True)\n" +
		fmt.Sprintf("plot.title('%s')\n", "foo") +
		"plot.show()\n"

	buf.WriteString(footer)

	_, err := fmt.Fprint(os.Stdout, buf.String())
	if err != nil {
		log.Fatal(err)
	}
}

func readZip(path string, data map[string]map[string]int) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("failed to open zip: %v", err)
	}

	cluster := filepath.Base(path)
	cluster = strings.TrimSuffix(cluster, ".zip")

	for _, file := range zr.File {
		f, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		if _, ok := data[file.Name]; !ok {
			data[file.Name] = make(map[string]int)
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			tabs := strings.Split(line, "\t")
			if len(tabs) == 2 {
				bytes, err := strconv.Atoi(tabs[0])
				if err != nil {
					return fmt.Errorf(err.Error())
				}

				user := strings.Replace(tabs[1], "/data/maildir/", "", -1)
				data[file.Name][fmt.Sprintf("%s/%s", cluster, user)] = bytes
			} else {
				log.Println(line)
			}
		}
	}

	return nil
}
