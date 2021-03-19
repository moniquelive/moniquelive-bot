package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gempir/go-twitch-irc/v2"
)

//go:embed .oauth
var oauth string

const (
	channel  = "moniquelive"
	username = "moniquelive_bot"
	//moniqueId = "4930146"
)

var commands []struct {
	Actions   []string `json:"actions"`
	Responses []string `json:"responses"`
}

func init() {
	reloadCommands()
}

func main() {
	roster := map[string]bool{}
	client := twitch.NewClient(username, oauth)

	client.OnConnect(func() {
		log.Println("*** OnConnect") // OnConnect attach callback to when a connection has been established
		client.Say(channel, "/color seagreen")
		client.Say(channel, "/me Tô na área!")
		client.Say(channel, "/slow 1")
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
		for _, command := range commands {
			matches := false
			for _, action := range command.Actions {
				matches = matches || strings.HasPrefix(message.Message, action)
			}
			if !matches {
				continue
			}
			for _, response := range command.Responses {
				parsed, err := parseTemplate(response, roster)
				if err != nil {
					log.Println("erro de template:", err)
					return
				}
				client.Say(channel, parsed)
			}
			return
		}

		// comando desconhecido...
		if strings.HasPrefix(message.Message, "!") {
			client.Say(channel, "/color firebrick")
			client.Say(channel, "/me ⁉ do que que você está falando?!")
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
					reloadCommands()
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
	var actions []string
	for _, command := range commands {
		actions = append(actions, command.Actions[0])
	}
	var vars struct {
		Roster   map[string]bool
		Commands string
	}
	vars.Roster = roster
	vars.Commands = strings.Join(actions, " ")

	tmpl, err := template.New("json").Parse(str)
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

func reloadCommands() {
	file, err := os.Open("commands.json")
	if err != nil {
		log.Fatalln("erro ao abrir commands.json:", err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&commands); err != nil {
		log.Fatalln("erro ao parsear commands.json:", err)
	}
}
