package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	queueName = "ms.tts"
	topicName = "create_tts"
)

var (
	amqpURL  = os.Getenv("RABBITMQ_URL")
	redisURL = os.Getenv("REDIS_URL")
	log      = logrus.WithField("package", "main")
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
	err = channel.QueueBind(queueName, topicName, "amq.topic", false, nil)
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
	const (
		writeWait  = 10 * time.Second    // Time allowed to write the file to the client.
		pongWait   = 30 * time.Second    // Time allowed to read the next pong message from the client.
		pingPeriod = (pongWait * 9) / 10 // Send pings to client with this period. Must be less than pongWait.
	)
	ws, _, err := dial()
	if err != nil {
		log.Println(err)
		return
	}
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		_ = ws.Close()
		log.Debugln("handle: deliveries channel closed")
		done <- struct{}{}
	}()

	log.Debugln("Listening...")
	for {
		select {
		case delivery := <-deliveries:
			if delivery.Body == nil {
				return
			}
			message := string(delivery.Body)
			if message == "" {
				log.Debugln("empty message. ignoring...")
				_ = delivery.Ack(false)
				break
			}
			log.Infoln("DELIVERY:", message)
			resp := tts(ws, message, "perola")
			if !resp.Payload.Success {
				log.Println("!Success:", resp.Payload.Reason)
				return
			}
			_ = delivery.Ack(false)
			// TODO: gerar evento de `tts_ready`
			// - jogar audio no redis
			// - publicar no topico `tts_ready`
			ffplay(resp.Payload.AudioURL)
		case <-pingTicker.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			log.Debugln("Ping...")
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Debugln("Ping:", err)
				return
			}
		}
	}
}
