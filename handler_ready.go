package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	guilds := s.State.Guilds

	log.Infof("%s Bot has started...running in %d Discord servers", self, len(guilds))

	// Register commands
	RegisterCommands(APPID)

	// Loop through all servers and update their component
	var server_objects []Server
	_ = db.Find(&server_objects)
	for _, server := range server_objects {
		server.Update()
	}

	// Find any Wordle stats that have been posted since the bot was down
	var wordle_channels []Wordle
	_ = db.Preload(clause.Associations).Find(&wordle_channels)
	for _, wordle := range wordle_channels {
		log.Debugf("Catching up on Wordle %s", wordle.ChannelID)
		wordle.CatchUp()
	}

	err := s.UpdateGameStatus(0, fmt.Sprintf("Wordle [v%s]", VERSION))
	if err != nil {
		log.Errorf("Error updating status: %s", err.Error())
	}

	log.Info("READY!")
}
