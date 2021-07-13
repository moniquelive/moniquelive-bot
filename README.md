# Moniquelive_Bot

## Links

- lib: https://github.com/gempir/go-twitch-irc
- oauth: https://twitchapps.com/tmi/
- limits (whisper, etc): https://dev.twitch.tv/limit-increase/

# Brainstorm

- [ ] limitar comando !urls para no maximo retornar 255 chars...
- [ ] comando !stats que mostra quantas vezes cada comando foi dado
- [ ] comando !m pode disparar o evento de WS para mostrar a musica no OBS
  - precisa fazer um refactoring para envio de AMQP ser menos burocratico
- [ ] comando !ragejs com contador
- [ ] comando !skip - abrir votação de x segundos para pular musica se maioria concordar
- [ ] twitch-bot_perola: investigar porque timeout na conexao nao derruba ela
- [ ] websocket: testar 2 clients ao mesmo tempo (`ch <- nil` vai zoar...)
- [ ] criar client-cli para postar/assinar filas do rabbitmq (Cobra SPF)
- [ ] comando !selfie para tocar video de auto-apresentacao (ola, sou a Monique...)
- [ ] comando !projeto do dia (!today/!hoje)
- [ ] (api da twitch) comando !uptime - informa quanto tempo a live está online
- [ ] (api da twitch) comando !schedule - mostra a agenda da twitch
- [ ] timers (alonga, hidrata, etc.)

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
- [x] websocket (des)conectando certinho?
- [x] comando !discord que exibe o telegram...
- [x] fs watcher não está funcionando
- [x] colocar notificação de musica no layer
- [x] comando !ban com sorteio de motivos aleatorios
- [x] localizar o horário do log do twitch-bot
- [x] extrair microserviço de dbus / spotify
- [x] comandos !uptime e !urls
