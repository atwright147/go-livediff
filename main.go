package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func colorDiff(text1, text2 string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	diffs = dmp.DiffCleanupEfficiency(diffs)

	var output strings.Builder

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			output.WriteString(color.GreenString(diff.Text))
		case diffmatchpatch.DiffDelete:
			output.WriteString(color.RedString(diff.Text))
		case diffmatchpatch.DiffEqual:
			output.WriteString(diff.Text)
		}
	}

	fmt.Println(output.String())
}

func watchFile(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Dir(absPath))
	if err != nil {
		log.Fatal(err)
	}

	previousContent, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Watching file: %s\n", absPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == absPath && event.Op&fsnotify.Write == fsnotify.Write {
				newContent, err := ioutil.ReadFile(absPath)
				if err != nil {
					log.Println("Error reading file:", err)
					continue
				}

				fmt.Printf("\n--- File changed at %s ---\n", time.Now().Format(time.RFC3339))
				colorDiff(string(previousContent), string(newContent))
				previousContent = newContent
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	watchFile(filePath)
}
