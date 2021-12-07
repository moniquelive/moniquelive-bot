package commands

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/moniquelive/moniquelive-bot/twitch/spotify"

	irc "github.com/gempir/go-twitch-irc/v2"
	"github.com/go-redis/redis"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
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
	ActionAdmin     map[string]bool
	ActionActions   map[string][]string
	actionAjuda     map[string]string
	actionHelp      map[string]string
}

const (
	redisUrlsKeyPrefix              = "twitch-bot:twitch_stats:urls:"
	redisSeenAtKeyPrefix            = "twitch-bot:twitch_stats:seen_at:"
	redisKeyCommandsPrefix          = "twitch-bot:twitch_stats:command:"
	skipMusicTopicName              = "spotify_music_skip"
	musicSkipPollName               = "twitch-bot:twitch:poll:skip_music"
	musicKeepPollName               = "twitch-bot:twitch:poll:keep_music"
	marqueeRedisKey                 = "twitch-bot:twitch:marquee:contents"
	appAccessTokenRedisKey          = "twitch-bot:twitch:app:access_token"
	userAccessTokenRedisKey         = "twitch-bot:twitch:user:access_token"
	userRefreshTokenRedisKey        = "twitch-bot:twitch:user:refresh_token"
	spotifyUserAccessTokenRedisKey  = "twitch-bot:spotify:user:access_token"
	spotifyUserRefreshTokenRedisKey = "twitch-bot:spotify:user:refresh_token"
	MoniqueliveID                   = "4930146"
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
var (
	//go:embed .oauth_client_id
	oauth_client_id string

	//go:embed .oauth_client_secret
	oauth_client_secret string

	//go:embed .spotify_client_id
	spotify_client_id string

	//go:embed .spotify_client_secret
	spotify_client_secret string
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
		log.Fatalln("Commands.init > Sem redis...")
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
		return fmt.Sprintf("%v: %v (sinônimos: %v)",
			action, help, strings.Join(c.ActionActions[action], ", "))
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
		return fmt.Sprintf("%v: %v (aliases: %v)",
			action, help, strings.Join(c.ActionActions[action], ", "))
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
	if username[0] == '@' {
		username = username[1:]
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
		urls = filterUrls(urls)
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

func filterUrls(urls []string) (uu []string) {
	for _, u := range urls {
		if strings.Contains(u, "open.spotify.com") {
			continue
		}
		uu = append(uu, u)
	}
	return
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

func (c Commands) Marquee(user *irc.User, cmdLine string) string {
	if !isAdmin(user) {
		return "Marquee > " + red.Get(marqueeRedisKey).Val()
	}
	if err := notifyAMQPTopic("marquee_updated", cmdLine); err != nil {
		log.Errorln("Marquee > notifyAMQPTopic:", err)
		return "Erro atualizando marquee: " + err.Error()
	}
	red.Set(marqueeRedisKey, cmdLine, 8*time.Hour)
	client, err := authHelix()
	if err != nil {
		return "Erro autenticando helix: " + err.Error()
	}
	channelInformation, err := client.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{MoniqueliveID},
	})
	if err != nil {
		return "Erro no GetChannelInformation: " + err.Error()
	}
	_, err = client.EditChannelInformation(&helix.EditChannelInformationParams{
		BroadcasterID:       MoniqueliveID,
		GameID:              channelInformation.Data.Channels[0].GameID,
		BroadcasterLanguage: channelInformation.Data.Channels[0].BroadcasterLanguage,
		Title:               cmdLine,
	})
	if err != nil {
		return "Erro no EditChannelInformation: " + err.Error()
	}
	return "Atualizando marquee: " + cmdLine
}

func (c Commands) SkipMusic(username string) string {
	username = strings.ToLower(username)
	red.SAdd(musicSkipPollName, username)
	skipMembers := red.SMembers(musicSkipPollName).Val()
	keepMembers := red.SMembers(musicKeepPollName).Val()
	skipVotes := len(skipMembers) - 1
	keepVotes := len(keepMembers) - 1
	if skipVotes-keepVotes > 5 {
		if err := notifyAMQPTopic(skipMusicTopicName, ""); err != nil {
			log.Errorln("Skip > notifyAMQPTopic:", err)
		}
		sort.Strings(skipMembers)
		return fmt.Sprintf("PULANDO!!!! 💃 (%v) X (%v)",
			strings.Join(Remove(".", skipMembers), ", "),
			strings.Join(Remove(".", keepMembers), ", "),
		)
	}
	return fmt.Sprintf("Aaaaa parciais: (vaza: %v X fica: %v)", skipVotes, keepVotes)
}

func (c Commands) KeepMusic(username string) string {
	username = strings.ToLower(username)
	red.SAdd(musicKeepPollName, username)
	keepVotes := len(red.SMembers(musicKeepPollName).Val()) - 1
	skipVotes := len(red.SMembers(musicSkipPollName).Val()) - 1
	return fmt.Sprintf("kumaPls parciais: (vaza: %v X fica: %v)", skipVotes, keepVotes)
}

func (c Commands) FollowAge(user *irc.User) string {
	userID := strings.ToLower(user.ID)
	if userID == "" {
		return c.Ajuda("followage")
	}
	client, err := authHelix()
	if err != nil {
		return "Erro autenticando helix: " + err.Error()
	}
	//
	// pega tempo de seguida
	//
	var resp *helix.UsersFollowsResponse
	resp, err = client.GetUsersFollows(&helix.UsersFollowsParams{
		FromID: userID,
		ToID:   MoniqueliveID,
	})
	if err != nil {
		return "Erro no GetUsersFollows: " + err.Error()
	}
	//
	// responde
	//
	if len(resp.Data.Follows) == 0 {
		return "Algo de errado não está certo " + user.Name + "..."
	}

	duration := time.Since(resp.Data.Follows[0].FollowedAt)
	return fmt.Sprintf("%s segue a monique live há %s dias", user.Name, FormatDuration(duration))
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
	c.ActionAdmin = make(map[string]bool)         // refresh action x admin map
	c.ActionActions = make(map[string][]string)   // refresh action x actions map
	c.actionAjuda = make(map[string]string)       // refresh action x Ajuda texts
	c.actionHelp = make(map[string]string)        // refresh action x Help texts
	for _, command := range c.Commands {
		responses := command.Responses
		extras := command.Extras
		actions := command.Actions
		logs := command.Logs
		ajuda := command.Ajuda
		help := command.Help
		for _, action := range command.Actions {
			c.ActionResponses[action] = responses
			c.ActionExtras[action] = extras
			c.ActionActions[action] = actions
			c.ActionLogs[action] = logs
			c.actionAjuda[action] = ajuda
			c.actionHelp[action] = help
		}
		if len(command.Actions) < 1 {
			continue
		}
	}
}

func (c *Commands) Actions() string {
	var sortedActions []string
	for _, command := range c.Commands {
		sortedActions = append(sortedActions, actionLabel(command.Actions))
	}
	sort.Strings(sortedActions)
	return strings.Join(sortedActions, " ")
}

func (c Commands) Hug(sender *irc.User, cmdLine string) string {
	if cmdLine == "" {
		return c.Ajuda("hug")
	}
	lowerCmdLine := strings.ToLower(cmdLine)
	if strings.HasPrefix(lowerCmdLine, "@") {
		lowerCmdLine = lowerCmdLine[1:]
	}
	if lowerCmdLine == strings.ToLower(sender.Name) {
		return fmt.Sprintf("♥ %s se auto-abraça 02Pat", sender.Name)
	}
	return fmt.Sprintf("♥ %s abraça %s 02Pat", sender.Name, cmdLine)
}

func (c Commands) SongRequest(user *irc.User, songUrl string) string {
	client, err := authSpotify()
	if err != nil {
		return "Erro autenticando spotify: " + err.Error()
	}

	parsedUrl, err := url.Parse(songUrl)
	if err != nil {
		return "Música não encontrada:" + err.Error()
	}
	split := strings.Split(parsedUrl.Path, "/")
	songId := split[len(split)-1]

	songInfo, err := client.GetSongInfo(songId)
	if err != nil {
		return "Música não encontrada:" + err.Error()
	}

	if err = client.EnqueueSong(songId); err != nil {
		return "Música não encontrada:" + err.Error()
	}
	return fmt.Sprintf("Enfileirando %q by %q - @%v",
		songInfo.Name,
		formattedArtists(songInfo),
		user.DisplayName)
}

func formattedArtists(info *spotify.SongInfoResponse) string {
	var artists []string
	for _, artist := range info.Artists {
		artists = append(artists, artist.Name)
	}
	return strings.Join(artists, ",")
}

func authHelix() (client *helix.Client, err error) {
	client, err = helix.NewClient(&helix.Options{
		ClientID:     oauth_client_id,
		ClientSecret: oauth_client_secret,
	})
	if err != nil {
		return nil, fmt.Errorf("erro no login: %v", err)
	}

	appAccessToken := red.Get(appAccessTokenRedisKey).Val()
	if appAccessToken == "" {
		var resp *helix.AppAccessTokenResponse
		resp, err = client.RequestAppAccessToken([]string{"channel:manage:broadcast"})

		if err != nil {
			return nil, fmt.Errorf("erro no access token: %v", err)
		}
		appAccessToken = resp.Data.AccessToken
		red.Set(appAccessTokenRedisKey, appAccessToken, time.Duration(resp.Data.ExpiresIn)*time.Second)
	}

	userAccessToken := red.Get(userAccessTokenRedisKey).Val()
	if userAccessToken == "" {
		userRefreshToken := red.Get(userRefreshTokenRedisKey).Val()
		var resp *helix.RefreshTokenResponse
		resp, err = client.RefreshUserAccessToken(userRefreshToken)
		if err != nil {
			return nil, fmt.Errorf("erro refreshing token: %v", err)
		}
		userAccessToken = resp.Data.AccessToken
		red.Set(userAccessTokenRedisKey, userAccessToken, time.Duration(resp.Data.ExpiresIn)*time.Second)
		red.Set(userRefreshTokenRedisKey, resp.Data.RefreshToken, 0)
	}

	client.SetAppAccessToken(appAccessToken)
	client.SetUserAccessToken(userAccessToken)
	return
}

func authSpotify() (client *spotify.Client, err error) {
	client, err = spotify.NewClient(&spotify.Options{
		ClientID:     spotify_client_id,
		ClientSecret: spotify_client_secret,
	})
	if err != nil {
		return nil, err
	}
	// get access token
	userAccessToken := red.Get(spotifyUserAccessTokenRedisKey).Val()
	if userAccessToken == "" {
		userRefreshToken := red.Get(spotifyUserRefreshTokenRedisKey).Val()
		var resp *spotify.RefreshTokenResponse
		resp, err = client.RefreshUserAccessToken(userRefreshToken)
		if err != nil {
			return nil, fmt.Errorf("erro refreshing token: %v", err)
		}
		userAccessToken = resp.AccessToken
		red.Set(spotifyUserAccessTokenRedisKey, userAccessToken, time.Duration(resp.ExpiresIn)*time.Second)
	}

	client.SetUserAccessToken(userAccessToken)
	return
}

func isAdmin(user *irc.User) bool {
	return user.ID == MoniqueliveID
}

func actionLabel(actions []string) string {
	count := 0
	for _, action := range actions {
		key := redisKeyCommandsPrefix + action[1:]
		str := red.Get(key).Val()
		if i, err := strconv.Atoi(str); err == nil {
			count += i
		}
	}
	if count == 0 {
		return actions[0]
	}
	return fmt.Sprintf("%v (%v)", actions[0], count)
}
