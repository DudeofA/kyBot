package main

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var commands = make(map[string]*discordgo.ApplicationCommand)
var currentCommands = make(map[string]*discordgo.ApplicationCommand)

func AddCommand(cmd *discordgo.ApplicationCommand) {
	commands[cmd.Name] = cmd
}

func RegisterCommands(appid string) {
	// Get current commands to check if new ones need to be added
	// Blank guildID means register commands globally across Discord
	guildID := ""
	if DEBUG {
		guildID = DEBUG_GUILD_ID

		// Remove any global commands from debug
		var currentGlobalCommands = make(map[string]*discordgo.ApplicationCommand)
		currentGlobalCommandArray, err := s.ApplicationCommands(appid, "")
		if err != nil {
			log.Errorln("Error fetching current apppliation commands for guild", err)
		}
		for _, command := range currentGlobalCommandArray {
			currentGlobalCommands[command.Name] = command
		}

		for _, command := range currentGlobalCommands {
			err := s.ApplicationCommandDelete(appid, "", command.ID)
			if err != nil {
				log.Errorln("Error deleting unused application command", err, command)
			} else {
				log.Debugf("Successfully unregistered global command: %s", command.Name)
			}
		}
	}

	currentCommandArray, err := s.ApplicationCommands(appid, guildID)
	if err != nil {
		log.Errorln("Error fetching current apppliation commands for guild", err)
	}

	if len(currentCommandArray) != 0 {
		log.Debug("Existing registered commands:")
		for _, command := range currentCommandArray {
			currentCommands[command.Name] = command
			log.Debugf("ID: %s | Name: %s", command.ID, command.Name)
		}
	}

	// Register any missing commands
	for _, command := range commands {
		if _, exists := currentCommands[command.Name]; !exists {
			command, err := s.ApplicationCommandCreate(appid, guildID, command)
			if err != nil {
				log.Errorln("Error creating application command", err, command)
			} else {
				log.Debugf("Registered command in [%s]: %s", guildID, command.Name)
			}
		}
	}

	// Unregister commands no longer used
	for _, command := range currentCommands {
		if _, exists := commands[command.Name]; !exists {
			err := s.ApplicationCommandDelete(appid, guildID, command.ID)
			if err != nil {
				log.Errorln("Error deleting unused application command", err, command)
			} else {
				log.Debugf("Unregistered command in [%s]: %s", guildID, command.Name)
			}
		}
	}
}
