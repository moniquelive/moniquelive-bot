package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/gorilla/websocket"
	"github.com/moniquelive/moniquelive-bot/internal/twitch"
)

type SongInfo struct {
	ImgUrl string `json:"imgUrl"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

func listenToDbus(ws *websocket.Conn, done chan struct{}, client *twitch.Twitch) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return err
	}
	defer conn.Close()

	const spotify = "org.mpris.MediaPlayer2.spotify"
	if err = conn.AddMatchSignal(
		dbus.WithMatchSender(spotify),
		dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		return err
	}

	dbusChan := make(chan *dbus.Signal, 10)
	conn.Signal(dbusChan)
	prevTrackID := ""
	// prevTrackIDTime := time.Now()
	for {
		select {
		case <-done:
			//log.Infoln("CAINDO FUERAAAAA!!!")
			return nil
		case v := <-dbusChan:
			data := v.Body[1].(map[string]dbus.Variant)
			metaData := data["Metadata"].Value()
			playbackStatus := data["PlaybackStatus"].Value().(string)
			if playbackStatus != "Playing" {
				//fmt.Println("*** Skipping:", playbackStatus)
				continue
			}
			songData := metaData.(map[string]dbus.Variant)
			trackID := songData["mpris:trackid"].Value().(string)
			if trackID == prevTrackID { //&& time.Now().Sub(prevTrackIDTime) < 2*time.Second {
				//fmt.Println("skipping for", trackID)
				continue
			}
			// prevTrackIDTime = time.Now()
			prevTrackID = trackID
			artist := songData["xesam:artist"].Value().([]string)[0]
			title := songData["xesam:title"].Value().(string)
			artUrl := songData["mpris:artUrl"].Value().(string)
			artUrl = strings.ReplaceAll(artUrl, "open.spotify.com", "i.scdn.co")

			err = ws.WriteJSON(SongInfo{
				ImgUrl: artUrl,
				Title:  title,
				Artist: artist,
			})
			// fmt.Println("WriteJSON:", err)
			if err != nil {
				return err
			}
			info := artist + " - " + title + " - " + artUrl
			log.Println(info)
			client.Say("/color Chocolate")
			client.Say("/me " + info)
		}
	}
}
