package commands

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var commands = make(map[string]*discordgo.ApplicationCommand)
var currentCommands = make(map[string]*discordgo.ApplicationCommand)

func AddCommand(cmd *discordgo.ApplicationCommand) {
	commands[cmd.Name] = cmd
}

func RegisterCommands(appid string, s *discordgo.Session) {
	// DELETE ALL COMMANDS IN ALL GUILDS
	// for _, guild := range s.State.Guilds {
	// 	commands, _ := s.ApplicationCommands(appid, guild.ID)
	// 	for _, command := range commands {
	// 		s.ApplicationCommandDelete(appid, guild.ID, command.ID)
	// 	}
	// }

	// Get current commands
	currentCommandArray, err := s.ApplicationCommands(appid, s.State.Guilds[1].ID)
	if err != nil {
		log.Errorln("Error fetching current apppliation commands for guild", err)
	}

	log.Debug("Current registered commands:")
	for _, command := range currentCommandArray {
		currentCommands[command.Name] = command
		log.Debugf("ID: %s | Name: %s", command.ID, command.Name)
	}

	// Register all commands
	for _, command := range commands {
		if _, exists := currentCommands[command.Name]; !exists {
			command, err := s.ApplicationCommandCreate(appid, s.State.Guilds[1].ID, command)
			if err != nil {
				log.Errorln("Error creating application command", err, command)
			} else {
				log.Debugf("Registered command in %s: %s", s.State.Guilds[1].Name, command.Name)
			}
		}
	}
}
