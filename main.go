package main

import (
	_ "embed"
	"log"
	"strconv"
	"strings"

	"github.com/gempir/go-twitch-irc/v2"
)

//go:embed .oauth
var oauth string

const (
	channel   = "moniquelive"
	username  = "moniquelive_bot"
	moniqueId = "4930146"
)

func main() {
	roster := map[string]bool{}
	client := twitch.NewClient(username, oauth)

	client.OnConnect(func() {
		log.Println("OnConnect") // OnConnect attach callback to when a connection has been established
		//var colors = [...]string{
		//	"Blue",
		//	"Coral",
		//	"DodgerBlue",
		//	"SpringGreen",
		//	"YellowGreen",
		//	"Green",
		//	"OrangeRed",
		//	"Red",
		//	"GoldenRod",
		//	"HotPink",
		//	"CadetBlue",
		//	"SeaGreen",
		//	"Chocolate",
		//	"BlueViolet",
		//	"Firebrick",
		//}
		//for _, color := range colors {
		//	client.Say(channel, "/color "+color)
		//	client.Say(channel, "/me "+color)
		//}
		client.Say(channel, "/color seagreen")
		client.Say(channel, "/me Tô na área!")
		client.Say(channel, "/slow 1")
		client.Say(channel, "/uniquechat")
	})

	client.OnWhisperMessage(func(message twitch.WhisperMessage) {
		log.Println("OnWhisperMessage: ", message) // OnWhisperMessage attach to new whisper
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Println("OnPrivateMessage: ", message) // OnPrivateMessage attach to new standard chat messages
	})

	client.OnClearChatMessage(func(message twitch.ClearChatMessage) {
		log.Println("OnClearChatMessage: ", message) // OnClearChatMessage attach to new messages such as timeouts
	})

	//client.OnClearMessage(func(message twitch.ClearMessage) {
	//	log.Println("OnClearMessage: ", message) // OnClearMessage attach when a single message is deleted
	//})

	//client.OnRoomStateMessage(func(message twitch.RoomStateMessage) {
	//	log.Println("OnRoomStateMessage: ", message) // OnRoomStateMessage attach to new messages such as submode enabled
	//})

	client.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
		log.Println("OnUserNoticeMessage: ", message) // OnUserNoticeMessage attach to new usernotice message such as sub, resub, and raids
	})

	//client.OnUserStateMessage(func(message twitch.UserStateMessage) {
	//	log.Println("OnUserStateMessage: ", message) // OnUserStateMessage attach to new userstate
	//})

	//client.OnGlobalUserStateMessage(func(message twitch.GlobalUserStateMessage) {
	//	log.Println("OnGlobalUserStateMessage: ", message) // OnGlobalUserStateMessage attach to new global user state
	//})

	//client.OnNoticeMessage(func(message twitch.NoticeMessage) {
	//	log.Println("OnNoticeMessage: ", message) // OnNoticeMessage attach to new notice message such as hosts
	//})

	client.OnUserJoinMessage(func(message twitch.UserJoinMessage) {
		roster[message.User] = true
	})

	client.OnUserPartMessage(func(message twitch.UserPartMessage) {
		delete(roster, message.User)
	})

	//client.OnReconnectMessage(func(message twitch.ReconnectMessage) {
	//	log.Println("OnReconnectMessage: ", message) // OnReconnectMessage attaches that is triggered whenever the twitch servers tell us to reconnect
	//})

	client.OnNamesMessage(func(message twitch.NamesMessage) {
		for _, user := range message.Users {
			roster[user] = true
		}
	})

	//// OnPingMessage attaches to PING message
	//client.OnPingMessage(func(message twitch.PingMessage) {
	//	log.Println("OnPingMessage: ", message)
	//})
	//
	//// OnPongMessage attaches to PONG message
	//client.OnPongMessage(func(message twitch.PongMessage) {
	//	log.Println("OnPongMessage: ", message)
	//})
	//
	//// OnUnsetMessage attaches to message types we currently don't support
	//client.OnUnsetMessage(func(message twitch.RawMessage) {
	//	log.Println("OnUnsetMessage: ", message)
	//})
	//
	//// OnPingSent attaches that's called whenever the client sends out a ping message
	//client.OnPingSent(func() {
	//	log.Println("OnPingSent")
	//})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Printf("%s (%v): %s\n", message.User.DisplayName, message.User.ID, message.Message)
		//if message.User.ID == moniqueId {
		//}
		if strings.HasPrefix(message.Message, "!roster") {
			client.Say(channel, "/color blue")
			client.Say(channel, "/me "+strconv.Itoa(len(roster)))
			//var users []string
			//for user := range roster {
			//	users = append(users, user)
			//}
			//msg := strings.Join(users, " ")
			//log.Println(">>>", msg)
		} else if strings.HasPrefix(message.Message, "!youtube") || strings.HasPrefix(message.Message, "!yt") {
			client.Say(channel, "/color yellowgreen")
			client.Say(channel, "/me playlist com os vídeos passados: https://www.youtube.com/playlist?list=PLyR9xTdgYCd9n4V5EK86aEqMd59QZeDD9")
		} else if strings.HasPrefix(message.Message, "!cybervox") || strings.HasPrefix(message.Message, "!vox") {
			client.Say(channel, "/color yellowgreen")
			client.Say(channel, "/me clique neste link no seu celular para falar com o robô: https://wa.me/5521982820415?text=%23mandaaudios&lang=pt_br")
		} else if strings.HasPrefix(message.Message, "!github") || strings.HasPrefix(message.Message, "!gh") {
			client.Say(channel, "/color yellowgreen")
			client.Say(channel, "/me clique neste link para ver nossos projetinhos: https://github.com/moniquelive")
		} else if strings.HasPrefix(message.Message, "!meme") {
			client.Say(channel, "/color yellowgreen")
			client.Say(channel, "/me clique neste link para ver nosso gerador de memes: https://meme.monique.dev")
		} else if strings.HasPrefix(message.Message, "!") {
			client.Say(channel, "/color firebrick")
			client.Say(channel, "/me do que que você está falando?!")
		}
	})

	client.Join(channel)

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}
