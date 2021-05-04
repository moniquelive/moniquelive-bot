package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/parnurzeal/gorequest"
)

//go:embed vox_client_id
var clientID string

//go:embed vox_client_secret
var clientSecret string

var accessToken string

type (
	ttsRequest struct {
		Emit    string            `json:"emit"`
		Payload ttsRequestPayload `json:"payload"`
	}
	ttsRequestPayload struct {
		Timestamp int64  `json:"timestamp"`
		Text      string `json:"text"`
		Voice     string `json:"voice"`
	}
	ttsResponse struct {
		Event   string             `json:"event"`
		Payload ttsResponsePayload `json:"payload"`
	}
	ttsResponsePayload struct {
		Success   bool   `json:"success"`
		Reason    string `json:"reason,omitempty"`
		AudioURL  string `json:"audio_url,omitempty"`
		Timestamp int64  `json:"timestamp,omitempty"`
	}
)

type (
	oauthRequest struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Audience     string `json:"audience"`
		GrantType    string `json:"grant_type"`
	}
	oauthResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
)

func fetchAccessToken(clientID string, clientSecret string) (oauthResponse, error) {
	var (
		request = oauthRequest{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Audience:     "https://api.cybervox.ai",
			GrantType:    "client_credentials",
		}
		response oauthResponse
	)

	log.Println("fetching access token...")
	res, body, errs := gorequest.New().Post("https://api.cybervox.ai/auth").
		Send(request).
		EndStruct(&response)
	if errs != nil {
		return oauthResponse{}, fmt.Errorf("gorequest(%s): %v", string(body), errs)
	}
	if res.StatusCode != http.StatusOK {
		return oauthResponse{}, fmt.Errorf("http.Status: %v", res.Status)
	}
	return response, nil
}

func getAccessToken(clientID, clientSecret string) (string, error) {
	if accessToken == "" {
		var response oauthResponse
		var err error
		if response, err = fetchAccessToken(clientID, clientSecret); err != nil {
			return "", err
		}
		accessToken = response.AccessToken
	}

	return accessToken, nil
}

func dial() (*websocket.Conn, *http.Response, error) {
	if clientID == "" || clientSecret == "" {
		return nil, nil, fmt.Errorf(`abort: check "CLIENT_ID" and "CLIENT_SECRET" envvars`)
	}
	var token string
	var err error
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	if token, err = getAccessToken(clientID, clientSecret); err != nil {
		return nil, nil, err
	}
	return websocket.DefaultDialer.Dial("wss://api.cybervox.ai/ws?access_token="+token, nil)
}

func tts(ws *websocket.Conn, responses <-chan ttsResponse, text, voice string) (response ttsResponse) {
	request := ttsRequest{
		Emit: "tts",
		Payload: ttsRequestPayload{
			Text:      text,
			Voice:     voice,
			Timestamp: time.Now().UnixNano(),
		},
	}
	_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := ws.WriteJSON(request); err != nil {
		log.Println(err)
		return
	}
	response = <-responses
	return
}

func ffplay(url string) {
	const urlPrefix = "https://api.cybervox.ai"
	cmd := fmt.Sprintf(`ffmpeg -i %q -filter:a "volume=5.0" -f wav - | ffplay -autoexit -nodisp -`,
		urlPrefix+url)
	fmt.Println("CMD:", cmd)
	if err := exec.Command("bash", "-c", cmd).Run(); err != nil {
		log.Println("ffplay:", err)
	}
}
