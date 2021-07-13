package commands_test

import (
	"testing"

	"github.com/moniquelive/moniquelive-bot/twitch/commands"
	"github.com/stretchr/testify/assert"
)

func TestWordWrap(t *testing.T) {
	var tt = []struct {
		name     string
		in       string
		size     int
		expected []string
	}{
		{"empty string, zero size", "", 0, []string{""}},
		{"empty string, 10 size", "", 10, []string{""}},
		{"a string that fits", "abc", 10, []string{"abc"}},
		{"a string that doesn't fit", "abc def ghi jkl", 10, []string{"abc def", "ghi jkl"}},
		{"edge case #1", "abcdefghijk l", 10, []string{"abcdefghijk", "l"}},
		{"edge case #2", "a bcdefghijkl", 10, []string{"a", "bcdefghijkl"}},
		{"real case", "moniquelive compartilhou: https://tour.golang.org/welcome/1 https://en.wikipedia.org/wiki/Test-driven_development https://hack.ainfosec.com/ https://github.com/moniquelive/iniciante - streamholics compartilhou: https://twitch.tv/alorenato https://twitch.tv/xtecna https://twitch.tv/adielseffrin https://twitch.tv/jpbrab0 https://twitch.tv/xtecna https://twitch.tv/kastr0walker https://twitch.tv/morgannadev https://twitch.tv/jpbrab0 https://twitch.tv/profbrunolopes https://twitch.tv/clauzinhando https://twitch.tv/pachicodes https://twitch.tv/adielseffrin https://twitch.tv/LadyDriveer https://twitch.tv/adielseffrin - acaverna compartilhou: https://twitch.tv/alorenato https://twitch.tv/xtecna https://twitch.tv/adielseffrin https://twitch.tv/jpbrab0 https://twitch.tv/xtecna https://twitch.tv/kastr0walker https://twitch.tv/morgannadev https://twitch.tv/jpbrab0 https://twitch.tv/profbrunolopes https://twitch.tv/clauzinhando https://twitch.tv/pachicodes https://twitch.tv/adielseffrin https://twitch.tv/adielseffrin - vivendoouexistindo compartilhou: https://discord.com/invite/cD7VJJZTnA - debora_666 compartilhou: https://mma.prnewswire.com/media/1438929/first_Logo.jpg?p=publish",
			500, []string{
			"moniquelive compartilhou: https://tour.golang.org/welcome/1 https://en.wikipedia.org/wiki/Test-driven_development https://hack.ainfosec.com/ https://github.com/moniquelive/iniciante - streamholics compartilhou: https://twitch.tv/alorenato https://twitch.tv/xtecna https://twitch.tv/adielseffrin https://twitch.tv/jpbrab0 https://twitch.tv/xtecna https://twitch.tv/kastr0walker https://twitch.tv/morgannadev https://twitch.tv/jpbrab0 https://twitch.tv/profbrunolopes https://twitch.tv/clauzinhando",
			"https://twitch.tv/pachicodes https://twitch.tv/adielseffrin https://twitch.tv/LadyDriveer https://twitch.tv/adielseffrin - acaverna compartilhou: https://twitch.tv/alorenato https://twitch.tv/xtecna https://twitch.tv/adielseffrin https://twitch.tv/jpbrab0 https://twitch.tv/xtecna https://twitch.tv/kastr0walker https://twitch.tv/morgannadev https://twitch.tv/jpbrab0 https://twitch.tv/profbrunolopes https://twitch.tv/clauzinhando https://twitch.tv/pachicodes https://twitch.tv/adielseffrin",
			"https://twitch.tv/adielseffrin - vivendoouexistindo compartilhou: https://discord.com/invite/cD7VJJZTnA - debora_666 compartilhou: https://mma.prnewswire.com/media/1438929/first_Logo.jpg?p=publish"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := commands.WordWrap(tc.in, tc.size)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIn(t *testing.T) {
	var tt = []struct {
		name     string
		elem     string
		slice    []string
		expected bool
	}{
		{"empty slice, always false #1", "", []string{}, false},
		{"empty slice, always false #2", "aaa", []string{}, false},
		{"existing element", "aaa", []string{"aaa"}, true},
		{"absent element", "", []string{"aaa"}, false},
		{"existing element multiples #1", "", []string{"aaa",""}, true},
		{"existing element multiples #2", "aaa", []string{"aaa",""}, true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := commands.In(tc.elem, tc.slice)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
