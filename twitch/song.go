package main

import "encoding/json"

type songInfo struct {
	ImgUrl  string `json:"imgUrl"`
	SongUrl string `json:"songUrl"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
	Length  int64  `json:"length"`
}

func parseSongInfo(bb []byte, songInfo *songInfo) error {
	return json.Unmarshal(bb, &songInfo)
}
