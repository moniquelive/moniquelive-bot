package main

import (
	"net/http"
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
	queueName = "ms.websocket"
	topicName = "spotify_music_updated"
)

var (
	amqpURL = os.Getenv("RABBITMQ_URL")
	log     = logrus.WithField("package", "main")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsHandler struct {
	writerChan chan []byte
	doneChan   chan struct{}
}

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
	defer log.Debugln("AMQP consumer shutdown.")

	var err error
	wsChan := make(chan []byte)
	wsDoneChan := make(chan struct{})

	//
	// Start websocket server
	//
	log.Println("Websocket Listening ...")
	router := http.NewServeMux()
	router.Handle("/ws", wsHandler{writerChan: wsChan, doneChan: wsDoneChan})

	// start server in a goroutine
	srv := &http.Server{Addr: ":9090", Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("ListenAndServe:", err)
		}
	}()

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

	deliveriesDoneChan := make(chan struct{})
	go handle(deliveries, wsChan, deliveriesDoneChan)

	// wait for interrupt signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-stopChan:
			// will close the deliveries channel
			err = channel.Cancel(tag, true)
			check(err)
			// will close websocket handler
			close(wsChan)

		// wait for go handle(...)
		case <-deliveriesDoneChan:
		case <-wsDoneChan:
			return
		}
	}
}

func handle(deliveries <-chan amqp.Delivery, ws chan<- []byte, done chan<- struct{}) {
	defer func() {
		log.Println("AMQP Handler: Exiting from deliveries handler")
		done <- struct{}{}
	}()
	log.Debugln("Listening...")
	for delivery := range deliveries {
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
		ws <- delivery.Body
		_ = delivery.Ack(false)
	}
}

func (ws wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const (
		readWait   = 5 * time.Second     // Time allowed to read the data from the client.
		pongWait   = 60 * time.Second    // Time allowed to read the next pong message from the client.
		pingPeriod = (pongWait * 9) / 10 // Send pings to client with this period. Must be less than pongWait.
	)
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("WS Handler: Exiting from wsHandler")
		pingTicker.Stop()
		ws.doneChan <- struct{}{}
	}()

	// Upgrade HTTP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln("wsHandler: Upgrade error:", err)
		return
	}
	defer func() { _ = conn.Close() }()

	log.Println("Websocket Conectado!")
	for {
		select {
		case body := <-ws.writerChan:
			if body == nil {
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, body)
			if err != nil {
				log.Errorln("ServeHTTP > writerChan:", err)
				return
			}
		case <-pingTicker.C:
			_ = conn.SetReadDeadline(time.Now().Add(readWait))
			_, _, _ = conn.ReadMessage()
		}
	}
}
