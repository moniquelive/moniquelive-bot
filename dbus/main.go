package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	topicName = "spotify_music_updated"
	redisKey  = "twitch-bot:dbus:song-info"
)

var (
	amqpURL  = os.Getenv("RABBITMQ_URL")
	redisURL = os.Getenv("REDIS_URL")
	log      = logrus.WithField("package", "main")
	red      *redis.Client
)

type SongInfo struct {
	ImgUrl  string `json:"imgUrl"`
	SongUrl string `json:"songUrl"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})
	logrus.SetLevel(logrus.TraceLevel) // sets log level
	red = redis.NewClient(&redis.Options{Addr: redisURL})
	if _, err := red.Ping().Result(); err != nil {
		log.Fatalln("dbus.init > Sem redis...")
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln("failed:", err)
	}
}

func main() {
	var err error

	conn, err := amqp.Dial(amqpURL)
	check(err)
	defer conn.Close()
	go func() { log.Debugf("closing: %s", <-conn.NotifyClose(make(chan *amqp.Error))) }()

	log.Debugln("got Connection, getting Channel")
	channel, err := conn.Channel()
	check(err)
	defer channel.Close()

	stopChan := make(chan os.Signal, 1)
	go func() {
		err := listenToDbus(channel, stopChan)
		check(err)
	}()

	// wait for interrupt signal
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	log.Debugln("AMQP consumer shutdown.")
}

func listenToDbus(channel *amqp.Channel, done chan os.Signal) error {
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

	log.Debugln("Listening...")
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
			songUrl := songData["xesam:url"].Value().(string)
			lengthInMillis := songData["mpris:length"].Value().(uint64) / 1000
			artUrl = strings.ReplaceAll(artUrl, "open.spotify.com", "i.scdn.co")

			songInfo := SongInfo{
				ImgUrl:  artUrl,
				SongUrl: songUrl,
				Title:   title,
				Artist:  artist,
			}
			//log.Debugln(songInfo)
			infoBytes, err := json.Marshal(songInfo)
			if err != nil {
				log.Errorln("listenToDbus > jsonMarshal:", err)
			}
			err = channel.Publish("amq.topic", topicName, false, false, amqp.Publishing{
				ContentType:     "application/json",
				ContentEncoding: "utf-8",
				DeliveryMode:    2,
				Expiration:      "60000",
				Body:            infoBytes,
			})
			if err != nil {
				log.Errorln("listenToDbus > channel.Publish:", err)
			}
			red.Set(redisKey, infoBytes, time.Duration(lengthInMillis)*time.Millisecond)
		}
	}
}
