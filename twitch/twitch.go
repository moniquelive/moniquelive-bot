package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/moniquelive/moniquelive-bot/twitch/commands"

	irc "github.com/gempir/go-twitch-irc/v2"
	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

const (
	moniqueID              = "4930146"
	channel                = "moniquelive"
	streamlabsID           = "105166207"
	TtsReward              = "e706421e-01f7-48fd-a4c6-4393d1ba4ec8"
	SpotifyReward          = "bf07c491-1ffb-4eb7-a7d8-5c9f2fe51818"
	redisKey               = "twitch-bot:dbus:song-info"
	twitchMessageTopicName = "twitch_message_delivered"
)

const (
	colorGreen = "\033[32m"
	colorWhite = "\033[30;47m"
	//colorYellow  = "\033[30;43m"
	colorRed = "\033[31m"
	//colorBlue    = "\033[34m"
	//colorMagenta = "\033[35m"
	colorCyan  = "\033[36m"
	colorReset = "\033[0m"
)

type Twitch struct {
	client      *irc.Client
	cmd         *commands.Commands
	rstr        *Roster
	amqpChannel *amqp.Channel
	player      *Player
}

type Player struct {
	red *redis.Client
}

func NewPlayer() (*Player, error) {
	redisURL := os.Getenv("REDIS_URL")
	red = redis.NewClient(&redis.Options{Addr: redisURL})
	if _, err := red.Ping().Result(); err != nil {
		return nil, fmt.Errorf("error pinging redis: %w", err)
	}
	return &Player{red: red}, nil
}

func (p Player) CurrentSong() string {
	infoBytes, err := red.Get(redisKey).Bytes()
	if err != nil {
		log.Errorln("CurrentSong.Get:", err)
		return "sem músicas no momento..."
	}
	var songInfo songInfo
	err = parseSongInfo(infoBytes, &songInfo)
	if err != nil {
		log.Errorln("CurrentSong.Unmarshal:", err)
		return "sem músicas no momento..."
	}
	return fmt.Sprintf("%v - %v - %v - %v",
		songInfo.Artist, songInfo.Title, songInfo.ImgUrl, songInfo.SongUrl)
}

func NewTwitch(username, oauth string, cmd *commands.Commands, amqpChannel *amqp.Channel) (*Twitch, error) {
	player, err := NewPlayer()
	if err != nil {
		return nil, err
	}
	client := irc.NewClient(username, oauth)
	t := &Twitch{
		client:      client,
		cmd:         cmd,
		player:      player,
		rstr:        NewRoster(),
		amqpChannel: amqpChannel,
	}
	client.OnConnect(func() {
		log.Println("*** OnConnect") // OnConnect attach callback to when a connection has been established
		t.Say("/color seagreen")
		t.Say("/me Tô na área!")
		// client.Say(channel, "/slow 1")
		t.Say("/uniquechat")
	})

	client.OnUserJoinMessage(func(message irc.UserJoinMessage) {
		publishTwitchMessage(t.amqpChannel, message.Raw)
		log.Println(colorGreen, "*** OnUserJoinMessage >>>", message.User, colorReset)
		t.rstr.AddUser(message.User)
	})

	client.OnUserPartMessage(func(message irc.UserPartMessage) {
		publishTwitchMessage(t.amqpChannel, message.Raw)
		log.Println(colorRed, "*** OnUserPartMessage <<<", message.User, colorReset)
		t.rstr.RemoveUser(message.User)
	})

	client.OnNamesMessage(func(message irc.NamesMessage) {
		publishTwitchMessage(t.amqpChannel, message.Raw)
		log.Println(colorWhite, "*** OnNamesMessage:", len(message.Users), colorReset)
		for _, user := range message.Users {
			t.rstr.AddUser(user)
		}
	})

	client.OnPrivateMessage(func(message irc.PrivateMessage) {
		//
		// atualiza contadores do !cmds
		//
		publishTwitchMessage(t.amqpChannel, message.Raw)
		//
		// ve se é o comando da pérola (channel points)
		//
		if rewardID, ok := message.Tags["custom-reward-id"]; ok && rewardID == TtsReward {
			err := t.amqpChannel.Publish("amq.topic", createTtsTopicName, false, false, amqp.Publishing{
				ContentType:     "text/plain",
				ContentEncoding: "utf-8",
				DeliveryMode:    2,
				Expiration:      "60000",
				Body:            []byte(message.Message),
			})
			if err != nil {
				log.Errorln("client.OnPrivateMessage > amqpChannel.Publish:", err)
			}
			return
		}
		//
		// ve se é o comando do Spotify (channel points)
		//
		if rewardID, ok := message.Tags["custom-reward-id"]; ok && rewardID == SpotifyReward {
			t.Say(cmd.SongRequest(&message.User, message.Message))
			return
		}

		// imprime log
		logWithColors(message.User.Name,
			fmt.Sprintf("%s (%v): %s", message.User.DisplayName, message.User.ID, message.Message))

		//
		// antivirus 🦠
		//
		if message.User.ID == streamlabsID {
			t.antivirus(message)
			return
		}
		// cai fora rápido se não for comando que começa com '!'
		if message.Message == "!" || message.Message[0] != '!' {
			return
		}
		// pula comandos marcados para ignorar
		for _, ignoredCommand := range cmd.IgnoredCommands {
			if strings.HasPrefix(message.Message, ignoredCommand) {
				return
			}
		}
		split := strings.Split(message.Message, " ")
		action := split[0]
		cmdLine := ""
		if len(split) > 1 {
			cmdLine = strings.Join(split[1:], " ")
		}
		//
		// verifica se é um comando privilegiado
		//
		admin, _ := cmd.ActionAdmin[action]
		if admin && message.User.ID != moniqueID {
			t.Say("/color firebrick")
			t.Say("Desculpa ai " + message.User.DisplayName + ", esse é só da Mo!")
			return
		}

		var (
			responses []string
			ok        bool
		)
		if responses, ok = cmd.ActionResponses[action]; !ok {
			// comando desconhecido...
			t.Say("/color firebrick")
			t.Say("/me não conheço esse: " + message.Message)
			return
		}

		extras, _ := cmd.ActionExtras[action] // parametros extras do comando
		for _, unparsedResponse := range responses {
			parsedResponse, err := t.parseTemplate(
				&message.User,
				unparsedResponse,
				cmdLine,
				extras)
			if err != nil {
				// TODO: tentar reproduzir esta condição de erro...
				split := strings.Split(err.Error(), ": ")
				errMsg := split[len(split)-1]
				errMsg = strings.ToUpper(errMsg[0:1]) + errMsg[1:]
				t.Say("/color red")
				t.Say("/me " + errMsg)
				return
			}
			for _, split := range strings.Split(parsedResponse, "\n") {
				t.Say(split)
			}
		}
		var logs []string
		if logs, ok = cmd.ActionLogs[action]; !ok || len(logs) == 0 {
			return
		}
		for _, unparsedLog := range logs {
			parsedLog, err := t.parseTemplate(
				&message.User,
				unparsedLog,
				cmdLine,
				[]string{})
			if err != nil {
				log.Println("erro de template:", err)
				return
			}
			fmt.Println(colorCyan, parsedLog, colorReset)
		}
	})

	client.Join(channel)
	return t, nil
}

func (t Twitch) antivirus(message irc.PrivateMessage) {
	rex := regexp.MustCompile(`Thank you for following (.*?)!`)
	if capture := rex.FindStringSubmatch(message.Message); capture != nil {
		nick := capture[1]
		if strings.HasPrefix(strings.ToLower(nick), "hoss00312_") ||
			strings.HasSuffix(strings.ToLower(nick), "_hoss00312") {
			t.Say("/ban " + nick)
			log.Println(colorRed, "!! TCHAU QUERIDO:", nick)
		}
	}
}

func publishTwitchMessage(amqpChannel *amqp.Channel, rawMessage string) {
	err := amqpChannel.Publish("amq.topic", twitchMessageTopicName, false, false, amqp.Publishing{
		ContentType:     "text/plain",
		ContentEncoding: "utf-8",
		DeliveryMode:    2,
		Expiration:      "5000",
		Body:            []byte(rawMessage),
	})
	if err != nil {
		log.Errorln("publishTwitchMessage > amqpChannel.Publish:", err)
	}
}

func (t Twitch) Say(msg string) {
	t.client.Say(channel, msg)
}

func (t Twitch) Connect() error {
	return t.client.Connect()
}

func (t Twitch) parseTemplate(
	user *irc.User,
	str,
	cmdLine string,
	extras []string,
) (_ string, err error) {
	var vars struct {
		Roster   Roster
		Player   Player
		Sender   *irc.User
		Commands string
		CmdLine  string
		Extras   []string
		Command  commands.Commands
	}
	vars.Sender = user
	vars.CmdLine = cmdLine
	vars.Extras = extras
	vars.Commands = t.cmd.Actions()
	vars.Command = *t.cmd
	vars.Player = *t.player
	vars.Roster = *t.rstr

	fns := template.FuncMap{
		"random": func(choices []string) string { return choices[rand.Intn(len(choices))] },
	}

	tmpl, err := template.New("json").Funcs(fns).Parse(str)
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

func logWithColors(userName, str string) {
	switch userName {
	case "acaverna", "streamlabs", "streamholics", "moniquelive_bot":
		log.Println(colorCyan, str, colorReset)
		return
	}
	log.Println(str)
}
