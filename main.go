package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	// "path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watchDirs := os.Getenv("WATCH_DIRS")
	chmod := os.Getenv("FILE_CHMOD")
	owner := os.Getenv("FILE_OWNER")
	group := os.Getenv("FILE_GROUP")

	if watchDirs == "" || chmod == "" || owner == "" || group == "" {
		log.Fatal("Environment variables WATCH_DIRS, FILE_CHMOD, FILE_OWNER, and FILE_GROUP must be set.")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("New file detected:", event.Name)
					applyPermissions(event.Name, chmod, owner, group)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, dir := range strings.Split(watchDirs, ",") {
		err = watcher.Add(strings.TrimSpace(dir))
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done
}

func applyPermissions(filePath, chmod, owner, group string) {
	if err := os.Chmod(filePath, parseChmod(chmod)); err != nil {
		log.Println("Failed to chmod file:", err)
	}
	chownCmd := exec.Command("chown", fmt.Sprintf("%s:%s", owner, group), filePath)
	if err := chownCmd.Run(); err != nil {
		log.Println("Failed to chown file:", err)
	}
}

func parseChmod(chmod string) os.FileMode {
	mode, err := strconv.ParseUint(chmod, 8, 32)
	if err != nil {
		log.Fatal("Invalid chmod value")
	}
	return os.FileMode(mode)
}
