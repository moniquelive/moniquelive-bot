# nc -l 9090
curl -i https://accounts.spotify.com/authorize\?client_id\=$(cat twitch/commands/.spotify_client_id)\&response_type\=code\&redirect_uri\=http://127.0.0.1:9090\&scope\=user-modify-playback-state

# cola "Location: (.*)" no browser

# pega code=(.*) no netcat

CODE= \
        curl -i -u $(cat twitch/commands/.spotify_client_id) \
                -H "Accept: application/json" \
                -d code=$CODE \
                -d redirect_uri=http://127.0.0.1:9090 \
                -d grant_type=authorization_code \
                -XPOST "https://accounts.spotify.com/api/token"

** LEMBRAR DE SETAR O TTL DO ACCESS_KEY **
