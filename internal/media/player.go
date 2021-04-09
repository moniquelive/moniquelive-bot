package media

import (
	"errors"
	"fmt"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/moniquelive/moniquelive-bot/internal/roster"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	spotify = "org.mpris.MediaPlayer2.spotify"
	vlc     = "org.mpris.MediaPlayer2.vlc"
)

type Player struct {
	roster *roster.Roster

	conn    *dbus.Conn
	client  *twitch.Client
	channel string

	skip map[string]bool
}

func New(roster *roster.Roster, client *twitch.Client, channel string) (*Player, func(), error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, nil, err
	}
	return &Player{
		roster:  roster,
		conn:    conn,
		client:  client,
		channel: channel,
	}, func() { _ = conn.Close() }, nil
}

func (p Player) SkipCurrentSong(user string) (string, error) {
	if p.skip != nil {
		(p.skip)[user] = true
		return "", nil
	}

	currentSong, err := p.CurrentSong()
	if err != nil {
		return "", err
	}
	p.createSkipPool(currentSong)

	return "", nil
}

func (p Player) CurrentSong() (string, error) {
	for _, dest := range []string{spotify, vlc} {
		obj := p.conn.Object(dest, "/org/mpris/MediaPlayer2")
		song, err := p.getMetadata(obj)
		if err != nil {
			continue
		}
		playbackStatus, err := p.getPlaybackStatus(obj)
		if err != nil {
			return "", err
		}
		status := playbackStatus.Value().(string)
		if strings.ToLower(status) != "playing" {
			continue
		}

		songData := song.Value().(map[string]dbus.Variant)
		switch dest {
		case spotify:
			artist := songData["xesam:artist"].Value().([]string)[0]
			title := songData["xesam:title"].Value().(string)
			artUrl := songData["mpris:artUrl"].Value().(string)
			artUrl = strings.ReplaceAll(artUrl, "open.spotify.com", "i.scdn.co")
			//rating := int(songData["xesam:autoRating"].Value().(float64) * 100)
			//url := songData["xesam:url"].Value().(string)
			return artist + " - " + title + " - " + artUrl, nil
		case vlc:
			return songData["vlc:nowplaying"].Value().(string), nil
		}
	}
	return "", errors.New("nenhum player tocando no momento")
}

func (p *Player) getMetadata(obj dbus.BusObject) (dbus.Variant, error) {
	song, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return dbus.Variant{}, err
	}
	return song, nil
}

func (p *Player) getPlaybackStatus(obj dbus.BusObject) (dbus.Variant, error) {
	playbackStatus, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		return dbus.Variant{}, err
	}
	return playbackStatus, nil
}

func (p *Player) CanGoNext(obj dbus.BusObject) (dbus.Variant, error) {
	canGoNextStatus, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.CanGoNext")
	if err != nil {
		return dbus.Variant{}, err
	}
	return canGoNextStatus, nil
}

func (p *Player) createSkipPool(currentSong string) {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	tenPercent := int(float32(len(p.roster.Keys())) * (0.1))
	increaseTime := int(10)

	splitSong := strings.Split(currentSong, " - ")
	if len(splitSong) > 2 {
		splitSong = splitSong[:len(splitSong)-1]
		currentSong = strings.Join(splitSong, " - ")
	}

	go func() {
		for {
			select {
			case <-done:
				if tenPercent == 0 || int(tenPercent-len(p.skip)) <= 0 {
					for _, dest := range []string{spotify, vlc} {
						obj := p.conn.Object(dest, "/org/mpris/MediaPlayer2")

						playbackStatus, err := p.getPlaybackStatus(obj)
						if err != nil {
							continue
						}

						status := playbackStatus.Value().(string)
						if strings.ToLower(status) != "playing" {
							continue
						}

						if dest == spotify {
							canGoNext, errNext := p.CanGoNext(obj)
							if errNext != nil {
								continue
							}

							goNext := canGoNext.Value().(bool)
							if goNext != true {
								continue
							}
						}

						// Skip song
						obj.Call("org.mpris.MediaPlayer2.Player.Next", 0)

						p.client.Say(p.channel, "/color blue")
						p.client.Say(p.channel, "/me Os loucos pularam a música "+currentSong)

						return
					}

					p.client.Say(p.channel, "/color red")
					p.client.Say(p.channel, "/me Não tem como passar pra frente loucos")
					return
				}

				p.client.Say(p.channel, "/color blue")
				p.client.Say(p.channel, "/me a música não foi pulada")

				return
			case <-ticker.C:
				p.client.Say(p.channel, "/color YellowGreen")
				p.client.Say(p.channel, fmt.Sprintf("/me Faltam %v loucos pra pular de música e %vs pra acabar a votação", int(tenPercent-len(p.skip)), increaseTime))

				increaseTime = int(increaseTime - 1)
			}
		}
	}()

	go func() {
		for {
			if tenPercent > 0 || int(tenPercent-len(p.skip)) > 0 {
				time.Sleep(10 * time.Second)
			}
			ticker.Stop()
			done <- true
		}
	}()
}
