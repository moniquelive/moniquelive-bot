# Moniquelive_Bot

## Links

- lib: https://github.com/gempir/go-twitch-irc
- oauth: https://twitchapps.com/tmi/
- limits (whisper, etc): https://dev.twitch.tv/limit-increase/

# Brainstorm

- [ ] webserver com api / layer de roster ou chamar obs websocket
- [ ] timeout (throttle) para comandos
- [ ] adicionar comando !help / !ajuda
- [ ] fator de correção no !roster (-3)
- [ ] comando !ban com sorteio de motivos aleatorios
- [ ] bookmark !blog, !twitter, !instagram, !projeto do dia (!today)
- [ ] timers (alonga, hidrata, etc.)

# DONE
- [x] comandos em _en_ e _pt-br_
- [x] /slow 1
- [x] /uniquechat
- [x] /me bot responses
- [x] bookmarks (!youtube/!yt,!cybervox/!vox,!github/!gh)
- [x] extrair comandos para .json
- [x] usar template para fazer comandos dinamicos
- [x] chamar funções GO pelo template (comando !rainbow)
- [x] "hot reload" config
- [x] hot reload for reals (real real)
- [x] refactoring: extract config.go
- [x] comando: !wiki
- [x] rectoring: formato do json, com ignored-commands e "logs" opcional por comando

