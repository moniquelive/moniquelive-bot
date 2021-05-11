# Moniquelive_Bot

## Links

- lib: https://github.com/gempir/go-twitch-irc
- oauth: https://twitchapps.com/tmi/
- limits (whisper, etc): https://dev.twitch.tv/limit-increase/

# Brainstorm

- [ ] criar client-cli para postar/assinar filas do rabbitmq (Cobra SPF)
- [ ] extrair microserviço de dbus / spotify
- [ ] comando !m pode disparar o evento de WS para mostrar a musica no OBS
- [ ] comando !selfie para tocar video de auto-apresentacao (ola, sou a Monique...)
- [ ] comando !ban com sorteio de motivos aleatorios
  - foi banid@ por curtir javascript
  - foi banid@ por perguntar em quais linguagens a strimer programa
  - foi banid@ por dar commit na master
  - foi banid@ por dar DELETE sem WHERE
  - tirou 1 no d20 e foi banido...
  - ban find ./moniquelive -name ${target} -delete
  - foi bando por esquecer o ;
- [ ] comando !projeto do dia (!today)
- [ ] comando !skip - abrir votação de x segundos para pular musica se maioria concordar
- [ ] (api da twitch) comando !uptime - informa quanto tempo a live está online
- [ ] (api da twitch) comando !schedule - mostra a agenda da twitch
- [ ] timers (alonga, hidrata, etc.)
- [ ] webserver com api / layer de roster ou chamar obs websocket
    - [ ] colocar notificação de musica no layer
- [ ] ? timeout (throttle) para comandos

# Pipe dream

- [ ] Mini-game na tela de #BRB

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
- [x] assinar evento de "status da playlist" para exibir a musica atual quando trocar
  - POC ok
  - refatorar parte do bot que envia mensagem, para enfileirar e evitar concorrencia
- [x] fazer index.js reconectar no websocket quando o bot parar
- [x] extrair microserviço de tts

