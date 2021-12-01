package spotify

import (
	"errors"
	"net/http"

	"github.com/parnurzeal/gorequest"
)

const (
	DefaultAPIBaseURL = "https://api.spotify.com/v1"
	RefreshTokenURL   = "https://accounts.spotify.com/api/token"
)

type (
	Client struct {
		opts *Options
	}
	Options struct {
		ClientID        string
		ClientSecret    string
		APIBaseURL      string
		UserAccessToken string
	}
	RefreshTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		ExpiresIn   int    `json:"expires_in"`
	}
)

type (
	ExternalID struct {
		Isrc string `json:"isrc"`
	}
	ExternalUrl struct {
		Spotify string `json:"spotify"`
	}
	Image struct {
		Height int    `json:"height"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
	}
	Album struct {
		AlbumType            string      `json:"album_type"`
		Artists              []Artist    `json:"artists"`
		ExternalUrls         ExternalUrl `json:"external_urls"`
		Href                 string      `json:"href"`
		Id                   string      `json:"id"`
		Images               []Image     `json:"images"`
		Name                 string      `json:"name"`
		ReleaseDate          string      `json:"release_date"`
		ReleaseDatePrecision string      `json:"release_date_precision"`
		TotalTracks          int         `json:"total_tracks"`
		Type                 string      `json:"type"`
		Uri                  string      `json:"uri"`
	}
	Artist struct {
		ExternalUrls ExternalUrl `json:"external_urls"`
		Href         string      `json:"href"`
		Id           string      `json:"id"`
		Name         string      `json:"name"`
		Type         string      `json:"type"`
		Uri          string      `json:"uri"`
	}
	SongInfoResponse struct {
		Album        Album       `json:"album"`
		Artists      []Artist    `json:"artists"`
		DiscNumber   int         `json:"disc_number"`
		DurationMs   int         `json:"duration_ms"`
		Explicit     bool        `json:"explicit"`
		ExternalIds  ExternalID  `json:"external_ids"`
		ExternalUrls ExternalUrl `json:"external_urls"`
		Href         string      `json:"href"`
		Id           string      `json:"id"`
		IsLocal      bool        `json:"is_local"`
		Name         string      `json:"name"`
		Popularity   int         `json:"popularity"`
		TrackNumber  int         `json:"track_number"`
		Type         string      `json:"type"`
		Uri          string      `json:"uri"`
	}
)

type refreshTokenRequestData struct {
	ClientID     string
	ClientSecret string
	GrantType    string
	RefreshToken string
}

func NewClient(options *Options) (client *Client, err error) {
	if options.ClientID == "" {
		return nil, errors.New("A client ID was not provided but is required")
	}
	if options.APIBaseURL == "" {
		options.APIBaseURL = DefaultAPIBaseURL
	}

	client = &Client{
		opts: options,
	}
	return client, nil
}

func (c *Client) RefreshUserAccessToken(refreshToken string) (resp *RefreshTokenResponse, err error) {
	opts := c.opts
	data := &refreshTokenRequestData{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}

	res, _, errs := gorequest.New().
		SetBasicAuth(opts.ClientID, opts.ClientSecret).
		Post(RefreshTokenURL).
		Set("Accept", "application/json").
		Send("client_id=" + data.ClientID).
		Send("client_secret=" + data.ClientSecret).
		Send("grant_type=" + data.GrantType).
		Send("refresh_token=" + data.RefreshToken).
		EndStruct(&resp)
	if errs != nil {
		return nil, errs[0]
	}
	if res.StatusCode != 200 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}
	return
}

func (c *Client) SetUserAccessToken(token string) {
	c.opts.UserAccessToken = token
}

func (c *Client) GetSongInfo(id string) (resp *SongInfoResponse, err error) {
	opts := c.opts

	res, _, errs := gorequest.New().
		Get(opts.APIBaseURL+"/tracks/"+id).
		Set("Authorization", "Bearer "+c.opts.UserAccessToken).
		EndStruct(&resp)
	if errs != nil {
		return nil, errs[0]
	}
	if res.StatusCode != 200 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}
	return
}

func (c *Client) EnqueueSong(id string) (err error) {
	opts := c.opts
	res, _, errs := gorequest.New().
		Post(opts.APIBaseURL+"/me/player/queue/").
		Set("Authorization", "Bearer "+c.opts.UserAccessToken).
		Query("uri=spotify:track:" + id).
		End()
	if errs != nil {
		return errs[0]
	}
	if res.StatusCode != 204 {
		return errors.New(http.StatusText(res.StatusCode))
	}
	return
}
