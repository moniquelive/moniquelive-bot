package roster

import (
	"log"
	"sort"

	"github.com/go-redis/redis"
)

type Roster map[string]bool

const redisSetKey = "moniquelive_bot:roster"
const redisChannel = "moniquelive_bot:notifications"

var red *redis.Client

func notify() {
	red.Publish(redisChannel, "updated")
}

func init() {
	red = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	if _, err := red.Ping().Result(); err != nil {
		log.Println("RosterType.init > Sem redis...")
		red = nil
	}
}

func New() *Roster {
	if red != nil {
		red.Del(redisSetKey)
		notify()
	}
	return &Roster{}
}

func (r *Roster) AddUser(userName string) {
	if red != nil {
		red.SAdd(redisSetKey, userName)
		notify()
	}
	(*r)[userName] = true
}

func (r *Roster) RemoveUser(userName string) {
	if red != nil {
		red.SRem(redisSetKey, userName)
		notify()
	}
	delete(*r, userName)
}

func (r Roster) Keys() []string {
	var result []string
	for k := range r {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
