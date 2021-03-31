package media

import (
	"errors"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	spotify = "org.mpris.MediaPlayer2.spotify"
	vlc     = "org.mpris.MediaPlayer2.vlc"
)

type Player struct {
	conn *dbus.Conn
}

func New() (*Player, func(), error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, nil, err
	}
	return &Player{conn: conn}, func() { conn.Close() }, nil
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
