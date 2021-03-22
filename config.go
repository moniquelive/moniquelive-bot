package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

type configType struct {
	IgnoredCommands []string `json:"ignored-commands"`
	Commands        []struct {
		Actions   []string `json:"actions"`
		Responses []string `json:"responses"`
		Logs      []string `json:"logs"`
	} `json:"commands"`
	actionResponses map[string][]string
	actionLogs      map[string][]string
	sortedActions   []string
}

func (c *configType) reload() {
	file, err := os.Open("commands.json")
	if err != nil {
		log.Fatalln("erro ao abrir commands.json:", err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(c); err != nil {
		log.Fatalln("erro ao parsear commands.json:", err)
	}
	c.refreshCache()
}

func (c *configType) refreshCache() {
	c.actionLogs = make(map[string][]string)      // refresh action x logs map
	c.actionResponses = make(map[string][]string) // refresh action x responses map
	c.sortedActions = nil                         // refresh sorted actions (for !commands)
	for _, command := range c.Commands {
		responses := command.Responses
		logs := command.Logs
		for _, action := range command.Actions {
			c.actionResponses[action] = responses
			c.actionLogs[action] = logs
		}
		if len(command.Actions) < 1 {
			continue
		}
		c.sortedActions = append(c.sortedActions, command.Actions[0])
	}
	sort.Strings(c.sortedActions)
}
