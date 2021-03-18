package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/gempir/go-twitch-irc/v2"
)

//go:embed .oauth
var oauth string

const (
	channel  = "moniquelive"
	username = "moniquelive_bot"
	//moniqueId = "4930146"
)

type commands []struct {
	Actions   []string `json:"actions"`
	Responses []string `json:"responses"`
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
		commands, err := parse(roster)
		if err != nil {
			log.Println("erro ao parsear commands.json:", err)
			return
		}
		for _, command := range commands {
			matches := false
			for _, action := range command.Actions {
				matches = matches || strings.HasPrefix(message.Message, action)
			}
			if matches {
				for _, response := range command.Responses {
					client.Say(channel, response)
				}
				return
			}
		}

		// comando desconhecido...
		if strings.HasPrefix(message.Message, "!") {
			client.Say(channel, "/color firebrick")
			client.Say(channel, "/me ⁉ do que que você está falando?!")
		}
	})

	client.Join(channel)

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

func parse(roster map[string]bool) (commands, error) {
	//
	// trata o .json como um arquivo texto opaco
	//
	jsonContents, err := os.ReadFile("commands.json")
	if err != nil {
		log.Fatalln("erro ao abrir commands.json:", err)
	}
	//
	// executa comandos do go template...
	//
	tmpl, err := template.New("json").Parse(string(jsonContents))
	if err != nil {
		return nil, err
	}
	var vars struct {
		Roster map[string]bool
	}
	vars.Roster = roster
	parsed := bytes.NewBufferString("")
	err = tmpl.Execute(parsed, vars)
	if err != nil {
		return nil, err
	}
	//
	// ... faz o parse do json como uma struct de go (commands)
	//
	var c commands
	if err := json.Unmarshal(parsed.Bytes(), &c); err != nil {
		return nil, err
	}
	return c, nil
}
