package twitch

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/moniquelive/moniquelive-bot/internal/commands"
	"github.com/moniquelive/moniquelive-bot/internal/media"
	"github.com/moniquelive/moniquelive-bot/internal/roster"

	irc "github.com/gempir/go-twitch-irc/v2"
)

const (
	channel = "moniquelive"
	//moniqueID = "4930146"
	streamlabsID = "105166207"
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

type twitch struct {
	client *irc.Client
	cmd    *commands.Commands
	player *media.Player
	rstr   *roster.Roster
}

func New(username, oauth string, cmd *commands.Commands) (*twitch, func(), error) {
	player, cancel, err := media.New()
	if err != nil {
		return nil, nil, err
	}
	client := irc.NewClient(username, oauth)
	t := &twitch{
		client: client,
		cmd:    cmd,
		player: player,
		rstr: roster.New(),
	}
	client.OnConnect(func() {
		log.Println("*** OnConnect") // OnConnect attach callback to when a connection has been established
		client.Say(channel, "/color seagreen")
		client.Say(channel, "/me Tô na área!")
		// client.Say(channel, "/slow 1")
		client.Say(channel, "/uniquechat")
	})

	client.OnUserJoinMessage(func(message irc.UserJoinMessage) {
		log.Println(colorGreen, "*** OnUserJoinMessage >>>", message.User, colorReset)
		t.rstr.AddUser(message.User)
	})

	client.OnUserPartMessage(func(message irc.UserPartMessage) {
		log.Println(colorRed, "*** OnUserPartMessage <<<", message.User, colorReset)
		t.rstr.RemoveUser(message.User)
	})

	client.OnNamesMessage(func(message irc.NamesMessage) {
		log.Println(colorWhite, "*** OnNamesMessage:", len(message.Users), colorReset)
		for _, user := range message.Users {
			t.rstr.AddUser(user)
		}
	})

	client.OnPrivateMessage(func(message irc.PrivateMessage) {
		setColorForUser(message.User.Name)
		log.Printf("%s (%v): %s%s\n", message.User.DisplayName, message.User.ID, message.Message, colorReset)
		//
		// deny list...
		//
		if message.User.ID == streamlabsID {
			return
		}
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
				parsedResponse, err := t.parseTemplate(unparsedResponse, cmdLine)
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
					parsedLog, err := t.parseTemplate(unparsedLog, cmdLine)
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

	return t, cancel, nil
}

func (t twitch) Connect() error {
	return t.client.Connect()
}

func (t twitch) parseTemplate(
	str string,
	cmdLine string,
) (_ string, err error) {
	var vars struct {
		Roster   roster.Roster
		Player   media.Player
		Commands string
		CmdLine  string
		Command  commands.Commands
	}
	vars.CmdLine = cmdLine
	vars.Commands = strings.Join(t.cmd.SortedActions, " ")
	vars.Command = *t.cmd
	vars.Player = *t.player
	vars.Roster = *t.rstr

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

func setColorForUser(userName string) {
	switch userName {
	case "acaverna", "streamlabs", "streamholics", "moniquelive_bot":
		log.Println(colorCyan)
	}
}
