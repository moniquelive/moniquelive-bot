package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gempir/go-twitch-irc/v2"
)

//go:embed .oauth
var oauth string

var config configType

const (
	channel  = "moniquelive"
	username = "moniquelive_bot"
	//moniqueId = "4930146"
)

func init() {
	config.reload()
}

func main() {
	roster := map[string]bool{}
	client := twitch.NewClient(username, oauth)

	client.OnConnect(func() {
		log.Println("*** OnConnect") // OnConnect attach callback to when a connection has been established
		client.Say(channel, "/color seagreen")
		client.Say(channel, "/me Tô na área!")
		// client.Say(channel, "/slow 1")
		client.Say(channel, "/uniquechat")
	})

	client.OnWhisperMessage(func(message twitch.WhisperMessage) { log.Println("OnWhisperMessage: ", message) })          // OnWhisperMessage attach to new whisper
	client.OnPrivateMessage(func(message twitch.PrivateMessage) { log.Println("OnPrivateMessage: ", message) })          // OnPrivateMessage attach to new standard chat messages
	client.OnClearChatMessage(func(message twitch.ClearChatMessage) { log.Println("OnClearChatMessage: ", message) })    // OnClearChatMessage attach to new messages such as timeouts
	client.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) { log.Println("OnUserNoticeMessage: ", message) }) // OnUserNoticeMessage attach to new usernotice message such as sub, resub, and raids

	client.OnUserJoinMessage(func(message twitch.UserJoinMessage) {
		log.Println("*** OnUserJoinMessage >>>", message.User)
		roster[message.User] = true
	})

	client.OnUserPartMessage(func(message twitch.UserPartMessage) {
		log.Println("*** OnUserPartMessage <<<", message.User)
		delete(roster, message.User)
	})

	client.OnNamesMessage(func(message twitch.NamesMessage) {
		log.Println("*** OnNamesMessage:", len(message.Users))
		for _, user := range message.Users {
			roster[user] = true
		}
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Printf("%s (%v): %s\n", message.User.DisplayName, message.User.ID, message.Message)
		//if message.User.ID == moniqueId {
		//}
		//var users []string
		//for user := range roster {
		//	users = append(users, user)
		//}
		//msg := strings.Join(users, " ")
		//log.Println(">>>", msg)
		// cai fora rápido se não for comando que começa com '!'
		if message.Message == "!" || message.Message[0] != '!' {
			return
		}
		responses, ok := config.actionResponses[message.Message]
		if ok {
			for _, response := range responses {
				parsed, err := parseTemplate(response, roster)
				if err != nil {
					log.Println("erro de template:", err)
					return
				}
				client.Say(channel, parsed)
			}
			if logs := config.actionLogs[message.Message]; len(logs) > 0 {
				for _, l := range logs {
					parsed, err := parseTemplate(l, roster)
					if err != nil {
						log.Println("erro de template:", err)
						return
					}
					fmt.Println(parsed)
				}
			}
			return
		}

		// pula comandos marcados para ignorar
		for _, ignoredCommand := range config.IgnoredCommands {
			if message.Message == ignoredCommand {
				return
			}
		}

		// comando desconhecido...
		if strings.HasPrefix(message.Message, "!") {
			client.Say(channel, "/color firebrick")
			client.Say(channel, "/me não conheço esse: "+message.Message)
		}
	})

	client.Join(channel)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watchCommandsFSChange(watcher)

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
					config.reload()
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

func parseTemplate(str string, roster map[string]bool) (_ string, err error) {
	fm := template.FuncMap{
		"keys": keys,
	}
	var vars struct {
		Roster   map[string]bool
		Commands string
	}
	vars.Roster = roster
	vars.Commands = strings.Join(config.sortedActions, " ")

	tmpl, err := template.New("json").Funcs(fm).Parse(str)
	if err != nil {
		return
	}
	parsed := bytes.NewBufferString("")
	err = tmpl.Execute(parsed, vars)
	if err != nil {
		return
	}
	return parsed.String(), nil
}
