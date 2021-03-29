package main

import (
	"log"

	"github.com/go-redis/redis"
)

type roster map[string]bool

const redisSetKey = "moniquelive_bot:roster"
const redisChannel = "moniquelive_bot:notifications"

var red *redis.Client

func notify() {
	red.Publish(redisChannel, "updated")
}

func init() {
	red = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	if _, err := red.Ping().Result(); err != nil {
		log.Println("roster.init > Sem redis...")
		red = nil
	}
}

func NewRoster() *roster {
	if red != nil {
		red.Del(redisSetKey)
		notify()
	}
	return &roster{}
}

func (r *roster) AddUser(userName string) {
	if red != nil {
		red.SAdd(redisSetKey, userName)
		notify()
	}
	(*r)[userName] = true
}

func (r *roster) RemoveUser(userName string) {
	if red != nil {
		red.SRem(redisSetKey, userName)
		notify()
	}
	delete(*r, userName)
}
