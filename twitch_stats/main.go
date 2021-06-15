package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	queueName              = "ms.twitch_stats"
	twitchMessageTopicName = "twitch_message_delivered"
)

var (
	amqpURL = os.Getenv("RABBITMQ_URL")
	log     = logrus.WithField("package", "main")
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})
	logrus.SetLevel(logrus.TraceLevel) // sets log level
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
	err = channel.QueueBind(queueName, twitchMessageTopicName, "amq.topic", false, nil)
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

	done := make(chan struct{})
	go handle(deliveries, done)

	// wait for interrupt signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stopChan:
		// will close the deliveries channel
		err = channel.Cancel(tag, true)
		check(err)

	// wait for go handle(...)
	case <-done:
		break
	}

	log.Debugln("AMQP consumer shutdown.")
}

func handle(deliveries <-chan amqp.Delivery, done chan<- struct{}) {
	defer func() {
		log.Debugln("handle: deliveries channel closed")
		done <- struct{}{}
	}()

	log.Debugln("Listening...")
	for delivery := range deliveries {
		if delivery.Body == nil {
			return
		}
		if len(delivery.Body) == 0 {
			log.Debugln("empty message. ignoring...")
			_ = delivery.Ack(false)
			break
		}
		body := string(delivery.Body)
		//log.Infoln("DELIVERY:", body)
		switch msg := twitch.ParseMessage(body).(type) {
		case *twitch.UserJoinMessage:
			log.Infoln("UserJoin: ", msg.User)
		case *twitch.UserPartMessage:
			log.Infoln("UserJoin: ", msg.User)
		case *twitch.NamesMessage:
			for _, user := range msg.Users {
				log.Info(user)
				log.Info(" ")
			}
			log.Infoln()
		case *twitch.PrivateMessage:
			log.Infof("PvtMessage: %v (%v): %v\n",
				msg.User.Name,
				msg.User.ID,
				msg.Message,
			)
		//default:
		//	log.Debugf("Desconhecido: %T\n", msg)
		}
		_ = delivery.Ack(false)
	}
}
