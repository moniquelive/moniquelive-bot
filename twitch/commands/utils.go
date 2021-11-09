package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

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

func FormatDuration(duration time.Duration) (format string) {
	plural := func(plural bool, prefix string) string {
		if !plural {
			return prefix
		}
		if strings.HasSuffix(prefix, "s") {
			return prefix + "es"
		}
		return prefix + "s"
	}
	for _, p := range []struct {
		millis time.Duration
		label  string
	}{
		{30 * 24 * time.Hour, "mes"},
		{24 * time.Hour, "dia"},
		{time.Hour, "hora"},
		{time.Minute, "minuto"},
	} {
		if duration >= p.millis {
			if format != "" {
				format += ", "
			}
			partial := duration / p.millis
			format += fmt.Sprintf("%d %s", partial, plural(partial > 1, p.label))
			duration -= partial * p.millis
		}
	}
	seconds := duration / time.Second
	if seconds > 0 {
		if format != "" {
			format += " e "
		}
		format += fmt.Sprintf("%d %s", seconds, plural(seconds > 1, "segundo"))
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
