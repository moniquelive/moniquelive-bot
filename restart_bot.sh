docker service update --force twitch-bot_twitch
docker service logs -n10 --raw -f --no-task-ids --no-trunc twitch-bot_twitch | cat

