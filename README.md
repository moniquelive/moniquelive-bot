# Moniquelive_Bot

## Links

- lib: https://github.com/gempir/go-twitch-irc
- oauth: https://twitchapps.com/tmi/
- limits (whisper, etc): https://dev.twitch.tv/limit-increase/

# Brainstorm

- [ ] comando !ban com sorteio de motivos aleatorios
- [ ] comando !projeto do dia (!today)
- [ ] comando !skip - abrir votação de x segundos para pular musica se maioria concordar
- [ ] assinar evento de "status da playlist" para exibir a musica atual quando trocar
- [ ] comando !uptime - informa quanto tempo a live está online (api da twitch)
- [ ] timers (alonga, hidrata, etc.)
- [ ] webserver com api / layer de roster ou chamar obs websocket
- [ ] ? timeout (throttle) para comandos

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
- [x] adicionar comando !help / !ajuda
- [x] bookmark !blog, !twitter, !instagram
- [x] log colorido
- [x] persistir lista de usuarios (roster) no redis, quando disponivel
- [x] tornar o type roster internal
- [x] renomear config.go para commands.go
- [x] adicionar comando !music com DBUS!
- [x] corrigir comando !help sem argumentos
- [x] comando !os com a versão do Linux
- [x] comando !pc/!spec com specs da maquina
- [x] comando !upside - https://en.wikipedia.org/wiki/Transformation_of_text#Upside-down_text
