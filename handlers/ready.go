package handlers

import (
	"fmt"
	"kyBot/commands"
	"kyBot/config"
	"kyBot/kyDB"
	"kyBot/status"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	guilds := s.State.Guilds

	log.Infof("%s Bot has started...running in %d Discord servers", self, len(guilds))

	// Register commands
	commands.RegisterCommands(config.APPID, s)

	kyDB.DB.AutoMigrate(&status.Server{}, &status.Wordle{}, &status.WordleStat{})

	// Loop through all servers and update their status
	var server_objects []status.Server
	_ = kyDB.DB.Find(&server_objects)
	for _, server := range server_objects {
		server.Update(s)
	}

	// Find any Wordle stats that have been posted since the bot was down
	var wordle_channels []status.Wordle
	_ = kyDB.DB.Preload(clause.Associations).Find(&wordle_channels)
	for _, wordle := range wordle_channels {
		wordle.CatchUp(s)
	}

	err := s.UpdateGameStatus(0, fmt.Sprintf("Wordle [v%s]", config.VERSION))
	if err != nil {
		log.Errorf("Error updating status: %s", err.Error())
	}

	log.Info("READY!")
}
