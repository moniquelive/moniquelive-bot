package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type configType struct {
	IgnoredCommands []string `json:"ignored-commands"`
	Commands        []struct {
		Actions   []string `json:"actions"`
		Responses []string `json:"responses"`
		Logs      []string `json:"logs"`
		Ajuda     string   `json:"ajuda"`
		Help      string   `json:"help"`
	} `json:"commands"`
	actionResponses map[string][]string
	actionLogs      map[string][]string
	actionAjuda     map[string]string
	actionHelp      map[string]string
	sortedActions   []string
}

func (c configType) Ajuda(cmdLine string) string {
	if cmdLine[0] != '!' {
		cmdLine = "!" + cmdLine
	}
	action := strings.Split(cmdLine, " ")[0]
	if help, ok := c.actionAjuda[action]; ok {
		return fmt.Sprintf("%v: %v", action, help)
	}
	return fmt.Sprintf("Comando %q não encontrado...", action)
}

func (c configType) Help(cmdLine string) string {
	if cmdLine[0] != '!' {
		cmdLine = "!" + cmdLine
	}
	action := strings.Split(cmdLine, " ")[0]
	if help, ok := c.actionHelp[action]; ok {
		return fmt.Sprintf("%v: %v", action, help)
	}
	return fmt.Sprintf("Help not found for %q...", action)
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
	c.actionAjuda = make(map[string]string)       // refresh action x Ajuda texts
	c.actionHelp = make(map[string]string)        // refresh action x Help texts
	for _, command := range c.Commands {
		responses := command.Responses
		logs := command.Logs
		ajuda := command.Ajuda
		help := command.Help
		for _, action := range command.Actions {
			c.actionResponses[action] = responses
			c.actionLogs[action] = logs
			c.actionAjuda[action] = ajuda
			c.actionHelp[action] = help
		}
		if len(command.Actions) < 1 {
			continue
		}
		c.sortedActions = append(c.sortedActions, command.Actions[0])
	}
	sort.Strings(c.sortedActions)
}
