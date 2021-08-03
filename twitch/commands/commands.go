package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Commands struct {
	IgnoredCommands []string `json:"ignored-commands"`
	Commands        []struct {
		Actions   []string `json:"actions"`
		Responses []string `json:"responses"`
		Logs      []string `json:"logs"`
		Extras    []string `json:"extras"`
		Ajuda     string   `json:"ajuda"`
		Help      string   `json:"help"`
	} `json:"commands"`
	ActionResponses map[string][]string
	ActionLogs      map[string][]string
	ActionExtras    map[string][]string
	SortedActions   []string
	actionAjuda     map[string]string
	actionHelp      map[string]string
}

const (
	redisUrlsKeyPrefix   = "twitch-bot:twitch_stats:urls:"
	redisSeenAtKeyPrefix = "twitch-bot:twitch_stats:seen_at:"
)

var (
	//https://en.wikipedia.org/wiki/Transformation_of_text#Upside-down_text
	lower    = []rune{'\u007A', '\u028E', '\u0078', '\u028D', '\u028C', '\u006E', '\u0287', '\u0073', '\u0279', '\u0062', '\u0064', '\u006F', '\u0075', '\u026F', '\u006C', '\u029E', '\u017F', '\u1D09', '\u0265', '\u0253', '\u025F', '\u01DD', '\u0070', '\u0254', '\u0071', '\u0250'}
	upper    = []rune{'\u005A', '\u2144', '\u0058', '\u004D', '\u039B', '\u0548', '\uA7B1', '\u0053', '\u1D1A', '\u10E2', '\u0500', '\u004F', '\u004E', '\uA7FD', '\u2142', '\uA4D8', '\u017F', '\u0049', '\u0048', '\u2141', '\u2132', '\u018E', '\u15E1', '\u0186', '\u15FA', '\u2200'}
	digits   = []rune{'\u0036', '\u0038', '\u3125', '\u0039', '\u100C', '\u07C8', '\u218B', '\u218A', '\u21C2', '\u0030'}
	punct    = []rune{'\u214B', '\u203E', '\u00BF', '\u00A1', '\u201E', '\u002C', '\u02D9', '\u0027', '\u061B'}
	charMap  = map[rune]rune{}
	log      = logrus.WithField("package", "commands")
	redisURL = os.Getenv("REDIS_URL")
	red      *redis.Client
)

func init() {
	fillMap('a', 'z', lower)
	fillMap('A', 'Z', upper)
	fillMap('0', '9', digits)
	for i, c := range "&_?!\"'.,;" {
		charMap[c] = punct[i]
	}
	charMap['('] = ')'
	charMap[')'] = '('
	charMap['{'] = '}'
	charMap['}'] = '{'
	charMap['['] = ']'
	charMap[']'] = '['

	red = redis.NewClient(&redis.Options{Addr: redisURL})
	if _, err := red.Ping().Result(); err != nil {
		log.Println("Commands.init > Sem redis...")
		red = nil
	}
}

func fillMap(from, to rune, slice []rune) {
	ll := int32(len(slice))
	for i := from; i <= to; i++ {
		charMap[i] = slice[ll-(i-from+1)]
	}
}

func (c Commands) Ajuda(cmdLine string) string {
	if cmdLine == "" {
		cmdLine = "ajuda"
	}
	if cmdLine[0] != '!' {
		cmdLine = "!" + cmdLine
	}
	action := strings.Split(cmdLine, " ")[0]
	if help, ok := c.actionAjuda[action]; ok {
		return fmt.Sprintf("%v: %v", action, help)
	}
	return fmt.Sprintf("Comando %q não encontrado...", action)
}

func (c Commands) Help(cmdLine string) string {
	if cmdLine == "" {
		cmdLine = "help"
	}
	if cmdLine[0] != '!' {
		cmdLine = "!" + cmdLine
	}
	action := strings.Split(cmdLine, " ")[0]
	if help, ok := c.actionHelp[action]; ok {
		return fmt.Sprintf("%v: %v", action, help)
	}
	return fmt.Sprintf("Help not found for %q...", action)
}

func (c Commands) Upside(cmdLine string) string {
	if cmdLine == "" {
		return c.Ajuda("upside")
	}
	result := ""
	for _, c := range cmdLine {
		if inv, ok := charMap[c]; ok {
			result = string(inv) + result
		} else {
			result = string(c) + result
		}
	}
	return result
}

func (c Commands) Ban(cmdLine string, extras []string) string {
	if cmdLine == "" {
		return c.Ajuda("ban")
	}
	randomExtra := extras[rand.Intn(len(extras))]
	return strings.ReplaceAll(randomExtra, "${target}", cmdLine)
}

func (c Commands) Urls(cmdLine string) []string {
	botList := []string{"acaverna", "streamholics"}
	username := strings.ToLower(cmdLine)
	if username == "" {
		username = "*"
	}
	allUsersRedisKeys := red.Keys(redisUrlsKeyPrefix + username).Val()
	if len(allUsersRedisKeys) == 0 {
		if username == "*" {
			username = "Ninguém"
		} else {
			username += " não"
		}
		return []string{username + " compartilhou urls ainda... :("}
	}
	var response []string
	for _, redisKey := range allUsersRedisKeys {
		split := strings.Split(redisKey, ":")
		username = split[len(split)-1]
		if In(username, botList) {
			continue
		}
		urls := red.LRange(redisUrlsKeyPrefix+username, 0, -1).Val()
		if len(urls) > 0 {
			response = append(response,
				username+" compartilhou: "+strings.Join(urls, " "))
		}
	}
	if len(response) == 0 {
		return []string{"Estranhaço... :S"}
	}
	return WordWrap(strings.Join(response, " - "), 500)
}

func (c Commands) Uptime(cmdLine string) string {
	username := strings.ToLower(cmdLine)
	if username == "" {
		return c.Ajuda("uptime")
	}
	unixtime := red.Get(redisSeenAtKeyPrefix + username).Val()
	if len(unixtime) == 0 {
		return username + " não tem horário de entrada... :("
	}
	uptime, err := strconv.ParseInt(unixtime, 10, 64)
	if err != nil {
		return "Tem algo de estranho que não está certo..."
	}
	t := time.Unix(uptime, 0)
	m := time.Since(t)
	return fmt.Sprintf("%s entrou dia %v ou seja, %v atrás",
		username,
		t.Format("02/01/2006 as 15:04:05"),
		m.Truncate(time.Second))
}

func (c Commands) Marquee(cmdLine string) string {
	amqpURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Errorln("Marquee > amqp.Dial:", err)
		return ""
	}
	defer conn.Close()
	go func() { <-conn.NotifyClose(make(chan *amqp.Error)) }()
	channel, err := conn.Channel()
	if err != nil {
		log.Errorln("Marquee > conn.Channel:", err)
		return ""
	}
	defer channel.Close()
	err = channel.Publish("amq.topic", "marquee_updated", false, false, amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		DeliveryMode:    2,
		Expiration:      "60000",
		Body:            []byte(cmdLine),
	})
	if err != nil {
		log.Errorln("Marquee > channel.Publish:", err)
		return ""
	}
	return "Atualizando marquee: " + cmdLine
}

func (c *Commands) Reload() {
	file, err := os.Open("./config/commands.json")
	if err != nil {
		log.Fatalln("erro ao abrir commands.json:", err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(c); err != nil {
		log.Fatalln("erro ao parsear commands.json:", err)
	}
	c.refreshCache()
}

func (c *Commands) refreshCache() {
	c.ActionLogs = make(map[string][]string)      // refresh action x logs map
	c.ActionResponses = make(map[string][]string) // refresh action x responses map
	c.ActionExtras = make(map[string][]string)    // refresh action x extras map
	c.SortedActions = nil                         // refresh sorted actions (for !commands)
	c.actionAjuda = make(map[string]string)       // refresh action x Ajuda texts
	c.actionHelp = make(map[string]string)        // refresh action x Help texts
	for _, command := range c.Commands {
		responses := command.Responses
		extras := command.Extras
		logs := command.Logs
		ajuda := command.Ajuda
		help := command.Help
		for _, action := range command.Actions {
			c.ActionResponses[action] = responses
			c.ActionExtras[action] = extras
			c.ActionLogs[action] = logs
			c.actionAjuda[action] = ajuda
			c.actionHelp[action] = help
		}
		if len(command.Actions) < 1 {
			continue
		}
		c.SortedActions = append(c.SortedActions, command.Actions[0])
	}
	sort.Strings(c.SortedActions)
}
