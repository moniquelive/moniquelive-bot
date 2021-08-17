package main

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"github.com/go-redis/redis"
)

const (
	defaultExpireDuration = 8 * time.Hour
	userRosterSet         = "twitch-bot:twitch_stats:user_roster"
	userDataKeySeenAt     = "twitch-bot:twitch_stats:seen_at:"
	userDataKeyURLs       = "twitch-bot:twitch_stats:urls:"
	userDataKeyCommands   = "twitch-bot:twitch_stats:command:"
)

var (
	red      *redis.Client
	redisURL = os.Getenv("REDIS_URL")
)

func init() {
	red = redis.NewClient(&redis.Options{Addr: redisURL})
	if _, err := red.Ping().Result(); err != nil {
		log.Println("RosterType.init > Sem redis...")
		red = nil
	}
}

func parseUserJoin(msg twitch.UserJoinMessage) {
	// adiciona usuário no conjunto de users
	// cria hashtable do usuário com campo de "seen_at time" se não existir
	userName := msg.User
	log.Infoln("UserJoin: ", userName)
	red.SAdd(userRosterSet, userName)
	setDefaultExpiration(userRosterSet)
	red.SetNX(userDataKeySeenAt+userName, time.Now().Unix(), defaultExpireDuration)
}

func parseUserPart(msg twitch.UserPartMessage) {
	// (nada por enquanto)
	log.Infoln("UserPart: ", msg.User)
}

func parseNames(msg twitch.NamesMessage) {
	// cria hashtable de inexistentes (vide OnUserJoin)
	for _, user := range msg.Users {
		parseUserJoin(twitch.UserJoinMessage{User: user})
	}
}

func parsePrivate(msg twitch.PrivateMessage) {
	log.Infof("PvtMessage: %v (%v): %v\n", msg.User.Name, msg.User.ID, msg.Message)

	parseHttps(msg)
	parseCommandsCounter(msg)
}

func parseCommandsCounter(msg twitch.PrivateMessage) {
	if !strings.HasPrefix(msg.Message, "!") {
		return
	}
	// !ola que tal -> split -> ["!ola", "que", "tal"] -> [0] -> !ola -> [1:] -> ola
	command := strings.Split(msg.Message, " ")[0][1:]
	red.Incr(userDataKeyCommands + command)
	setDefaultExpiration(userDataKeyCommands + command)
}

func parseHttps(msg twitch.PrivateMessage) {
	// regexp:
	//  adiciona url em lista de urls para usuário
	//  conta quantas vezes demos cada comando...
	re := regexp.MustCompile(`^https?://`)
	var urls []string
	for _, s := range strings.Split(msg.Message, " ") {
		if re.MatchString(s) {
			urls = append(urls, s)
		}
	}
	red.LPush(userDataKeyURLs+msg.User.Name, urls)
	setDefaultExpiration(userDataKeyURLs + msg.User.Name)
}

func setDefaultExpiration(key string) {
	if red.TTL(key).Val() == -1*time.Second {
		red.Expire(key, defaultExpireDuration)
	}
}
