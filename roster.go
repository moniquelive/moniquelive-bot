package main

import (
	"github.com/go-redis/redis"
	"log"
)

type roster map[string]bool

const redisSetKey = "moniquelive_bot:roster"

var red *redis.Client

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
	}
	return &roster{}
}

func (r *roster) AddUser(userName string) {
	if red != nil {
		red.SAdd(redisSetKey, userName)
	}
	(*r)[userName] = true
}

func (r *roster) RemoveUser(userName string) {
	if red != nil {
		red.SRem(redisSetKey, userName)
	}
	delete(*r, userName)
}
