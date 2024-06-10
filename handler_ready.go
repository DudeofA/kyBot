package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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

	err := s.UpdateGameStatus(0, fmt.Sprintf("v%s", VERSION))
	if err != nil {
		log.Errorf("Error updating status: %s", err.Error())
	}

	log.Info("READY!")
}
