package commands

import (
	"os"
	"strings"

	"github.com/streadway/amqp"
)

func WordWrap(str string, size int) (retval []string) {
	if len(str) <= size {
		return []string{str}
	}
	splits := strings.Split(str, " ")
	acc := ""
	for _, split := range splits {
		if len(acc+split) > size {
			if trim := strings.TrimSpace(acc); len(trim) > 0 {
				retval = append(retval, trim[:])
			}
			acc = split
		} else {
			acc += " " + split
		}
	}
	if trim := strings.TrimSpace(acc); len(trim) > 0 {
		retval = append(retval, trim[:])
	}
	return
}

func In(elem string, slice []string) bool {
	for _, e := range slice {
		if elem == e {
			return true
		}
	}
	return false
}

func Remove(elem string, slice []string) (newSlice []string) {
	newSlice = []string{}
	for _, e := range slice {
		if elem != e {
			newSlice = append(newSlice, e)
		}
	}
	return
}

func notifyAMQPTopic(topicName, body string) error {
	amqpURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	go func() { <-conn.NotifyClose(make(chan *amqp.Error)) }()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	err = channel.Publish("amq.topic", topicName, false, false, amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		DeliveryMode:    2,
		Expiration:      "60000",
		Body:            []byte(body),
	})
	if err != nil {
		return err
	}
	return nil
}
