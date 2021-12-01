package commands_test

import (
	"log"
	"testing"

	irc "github.com/gempir/go-twitch-irc/v2"
	"github.com/moniquelive/moniquelive-bot/twitch/commands"
)

func TestSongRequest(t *testing.T) {
	c := commands.Commands{}
	ret := c.SongRequest(&irc.User{
		DisplayName: "cyberama",
	}, "https://open.spotify.com/track/6OufwUcCqo81guU2jAlDVP?si=2a9566a0f7dc4f50")
	log.Println(ret)
}
