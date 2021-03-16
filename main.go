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
	channel   = "moniquelive"
	username  = "moniquelive_bot"
	moniqueId = "4930146"
)

var commands []struct {
	Actions  []string `json:"actions"`
	Color    string   `json:"color"`
	Response string   `json:"response"`
}

func init() {
	file, err := os.Open("commands.json")
	if err != nil {
		log.Fatalln("erro ao abrir commands.json:", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&commands); err != nil {
		log.Fatalln("erro ao parsear commands.json:", err)
	}
}

func main() {
	roster := map[string]bool{}
	client := twitch.NewClient(username, oauth)

	client.OnConnect(func() {
		log.Println("OnConnect") // OnConnect attach callback to when a connection has been established
		//rainbow(client)
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
		roster[message.User] = true
	})

	client.OnUserPartMessage(func(message twitch.UserPartMessage) {
		delete(roster, message.User)
	})

	client.OnNamesMessage(func(message twitch.NamesMessage) {
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
			if matches {
				response, err := parse(roster, command.Response)
				if err != nil {
					log.Printf("erro ao parsear command %q: %v", command.Response, err)
					return
				}
				client.Say(channel, "/color "+command.Color)
				client.Say(channel, response)
				return
			}
		}

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

func parse(roster map[string]bool, response string) (string, error) {
	var vars struct {
		Roster map[string]bool
	}
	vars.Roster = roster
	tmpl, err := template.New("command").Parse(response)
	if err != nil {
		return "", err
	}
	parsed := bytes.NewBufferString("")
	err = tmpl.Execute(parsed, vars)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func rainbow(client *twitch.Client) {
	var colors = [...]string{
		"Blue",
		"Coral",
		"DodgerBlue",
		"SpringGreen",
		"YellowGreen",
		"Green",
		"OrangeRed",
		"Red",
		"GoldenRod",
		"HotPink",
		"CadetBlue",
		"SeaGreen",
		"Chocolate",
		"BlueViolet",
		"Firebrick",
	}
	for _, color := range colors {
		client.Say(channel, "/color "+color)
		client.Say(channel, "/me "+color)
	}
}
