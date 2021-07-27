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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	musicUpdatedTopicName = "spotify_music_updated"
	skipMusicTopicName    = "spotify_music_skip"
	redisKey              = "twitch-bot:dbus:song-info"
	queueName             = "ms.dbus"
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

	dbusDoneChan := make(chan struct{})
	mqDoneChan := make(chan struct{})

	{
		log.Debugln("getting sending Channel")
		sendingMQChannel, err := conn.Channel()
		check(err)
		defer sendingMQChannel.Close()

		go func() {
			err := listenToDbus(sendingMQChannel, dbusDoneChan)
			check(err)
		}()
	}

	{
		log.Debugln("getting receiving Channel")
		channel, err := conn.Channel()
		check(err)
		defer channel.Close()

		log.Debugf("declaring Queue %q", queueName)
		queue, err := channel.QueueDeclare(
			queueName, // name of the queue
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // noWait
			nil,       // arguments
		)
		check(err)

		log.Debugf("binding Queue %q to amq.topic", queueName)
		err = channel.QueueBind(queueName, skipMusicTopicName, "amq.topic", false, nil)
		check(err)

		log.Debugln("Setting QoS")
		err = channel.Qos(1, 0, true)
		check(err)

		log.Debugf("declared Queue (%q %d messages, %d consumers)", queue.Name, queue.Messages, queue.Consumers)

		tag := uuid.NewString()
		log.Debugf("starting Consume (tag:%q)", tag)
		deliveries, err := channel.Consume(
			queue.Name, // name
			tag,        // consumerTag,
			false,      // noAck
			false,      // exclusive
			false,      // noLocal
			false,      // noWait
			nil,        // arguments
		)
		check(err)

		go handle(deliveries, mqDoneChan)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	dbusDoneChan <- struct{}{}
	mqDoneChan <- struct{}{}

	log.Debugln("AMQP consumer shutdown.")
}

func listenToDbus(channel *amqp.Channel, done <-chan struct{}) error {
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

	log.Debugln("DBus Listening...")
	// prevTrackIDTime := time.Now()
	for {
		select {
		case <-done:
			//log.Debugln("DBUS CAINDO FUERAAAAA!!!")
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
			err = channel.Publish("amq.topic", musicUpdatedTopicName, false, false, amqp.Publishing{
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

func handle(deliveries <-chan amqp.Delivery, done <-chan struct{}) {
	defer log.Println("AMQP Handler: Exiting from deliveries handler")

	dbusConn, err := dbus.ConnectSessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return
	}
	defer dbusConn.Close()

	const spotify = "org.mpris.MediaPlayer2.spotify"

	log.Debugln("MQ Listening...")
	for {
		select {
		case <-done:
			//log.Debugln("MQ CAINDO FUERAAAAA!!!")
			return
		case delivery := <-deliveries:
			if delivery.Body != nil && len(delivery.Body) > 0 {
				log.Infoln("SkipMusic Body (??):", string(delivery.Body))
			}

			switch delivery.RoutingKey {
			case skipMusicTopicName:
				dbusConn.
				Object(spotify, "/org/mpris/MediaPlayer2").
				Call("org.mpris.MediaPlayer2.Player.Next", 0)
			}

			_ = delivery.Ack(false)
		}
	}
}
