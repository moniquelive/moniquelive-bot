package main

import (
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func NewWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	go watchCommandsFSChange(watcher)

	err = watcher.Add("./")
	if err != nil {
		log.Fatalln(err)
	}
	err = watcher.Add("./commands.json")
	if err != nil {
		log.Fatalln(err)
	}
}

func watchCommandsFSChange(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("watchCommandsFSChange > events quit")
				return
			}
			//log.Println("watchCommandsFSChange > event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("watchCommandsFSChange > modified file:", event.Name)
				time.Sleep(1 * time.Second)
				cmd.Reload()
			}
			if event.Op&fsnotify.Create == fsnotify.Create && strings.HasSuffix(event.Name, "commands.json") {
				log.Println("watchCommandsFSChange > re-watching:", event.Name)
				if err := watcher.Add("./commands.json"); err != nil {
					log.Println("watchCommandsFSChange > watcher.Add:", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("watchCommandsFSChange > errors quit")
				return
			}
			log.Println("watchCommandsFSChange > error:", err)
		}
	}
}
