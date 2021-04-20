package main

import (
	_ "embed"
	"net/http"
	"strings"
	"time"

	"github.com/moniquelive/moniquelive-bot/internal/commands"
	"github.com/moniquelive/moniquelive-bot/internal/twitch"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

const username = "moniquelive_bot"

//go:embed .oauth
var oauth string

var cmd commands.Commands

var log = logrus.WithField("package", "main")

func init() {
	cmd.Reload()
}

func main() {
	client, cancel, err := twitch.New(username, oauth, &cmd)
	if err != nil {
		log.Panicln("twitch.New(): ", err)
	}
	defer cancel()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watchCommandsFSChange(watcher)

	//
	// Start websocket server
	//
	log.Println("Listening ...")
	router := http.NewServeMux()
	router.Handle("/ws", http.HandlerFunc(wsHandler))

	// start server in a goroutine
	srv := &http.Server{Addr: ":9090", Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	//
	// Start IRC loop
	//
	err = client.Connect()
	if err != nil {
		panic(err)
	}
}

func watchCommandsFSChange(watcher *fsnotify.Watcher) {
	go func() {
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
	}()

	if err := watcher.Add("./"); err != nil {
		log.Fatalln(err)
	}

	if err := watcher.Add("./commands.json"); err != nil {
		log.Fatalln(err)
	}
}
