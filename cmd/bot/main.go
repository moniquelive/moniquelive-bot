package main

import (
	"github.com/moniquelive/moniquelive-bot/internal/commands"
	"github.com/moniquelive/moniquelive-bot/internal/media"
	"github.com/moniquelive/moniquelive-bot/internal/roster"

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

var cmd commands.Commands

const (
	channel  = "moniquelive"
	username = "moniquelive_bot"
	//moniqueId = "4930146"
)

const (
	colorGreen   = "\033[32m"
	colorWhite   = "\033[30;47m"
	colorYellow  = "\033[30;43m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
)

func init() {
	cmd.Reload()
}

func main() {
	rstr := roster.New()
	client := twitch.NewClient(username, oauth)
	player, cancel, err := media.New()
	if err != nil {
		log.Panicln("media.New(): ", err)
	}
	defer cancel()

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
		log.Println(colorGreen, "*** OnUserJoinMessage >>>", message.User, colorReset)
		rstr.AddUser(message.User)
	})

	client.OnUserPartMessage(func(message twitch.UserPartMessage) {
		log.Println(colorRed, "*** OnUserPartMessage <<<", message.User, colorReset)
		rstr.RemoveUser(message.User)
	})

	client.OnNamesMessage(func(message twitch.NamesMessage) {
		log.Println(colorWhite, "*** OnNamesMessage:", len(message.Users), colorReset)
		for _, user := range message.Users {
			rstr.AddUser(user)
		}
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		setColorForUser(message.User.Name)
		log.Printf("%s (%v): %s%s\n", message.User.DisplayName, message.User.ID, message.Message, colorReset)
		//if message.User.ID == moniqueId {
		//}
		// cai fora rápido se não for comando que começa com '!'
		if message.Message == "!" || message.Message[0] != '!' {
			return
		}
		split := strings.Split(message.Message, " ")
		action := split[0]
		cmdLine := ""
		if len(split) > 1 {
			cmdLine = strings.Join(split[1:], " ")
		}
		responses, ok := cmd.ActionResponses[action]
		if ok {
			for _, unparsedResponse := range responses {
				parsedResponse, err := parseTemplate(unparsedResponse, rstr, cmdLine, player)
				if err != nil {
					// TODO: SE LIVRAR DESTE LIXOOOOOOOOO
					split := strings.Split(err.Error(), ": ")
					errMsg := split[len(split)-1]
					errMsg = strings.ToUpper(errMsg[0:1]) + errMsg[1:]
					client.Say(channel, "/color red")
					client.Say(channel, "/me "+errMsg)
					return
				}
				client.Say(channel, parsedResponse)
			}
			if logs := cmd.ActionLogs[action]; len(logs) > 0 {
				for _, unparsedLog := range logs {
					parsedLog, err := parseTemplate(unparsedLog, rstr, cmdLine, player)
					if err != nil {
						log.Println("erro de template:", err)
						return
					}
					fmt.Println(parsedLog)
				}
			}
			return
		}

		// pula comandos marcados para ignorar
		for _, ignoredCommand := range cmd.IgnoredCommands {
			if strings.HasPrefix(message.Message, ignoredCommand) {
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

func setColorForUser(userName string) {
	switch userName {
	case "acaverna", "streamlabs", "streamholics", "moniquelive_bot":
		log.Println(colorCyan)
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

func parseTemplate(
	str string,
	r *roster.Roster,
	cmdLine string,
	p *media.Player,
) (_ string, err error) {
	var vars struct {
		Roster   roster.Roster
		Player   media.Player
		Commands string
		CmdLine  string
		Command  commands.Commands
	}
	vars.Roster = *r
	vars.Player = *p
	vars.Commands = strings.Join(cmd.SortedActions, " ")
	vars.CmdLine = cmdLine
	vars.Command = cmd

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
