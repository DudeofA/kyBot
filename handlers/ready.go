package handlers

import (
	"kyBot/commands"
	"kyBot/config"
	"kyBot/kyDB"
	"kyBot/servers"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	guilds := s.State.Guilds

	log.Infof("%s Bot has started...running in %d Discord servers", self, len(guilds))

	// Register commands
	commands.RegisterCommands(config.APPID, s)

	// Loop through all Minecraft servers and update their status
	kyDB.DB.AutoMigrate(&servers.Server{})
	var server_objects []servers.Server
	_ = kyDB.DB.Find(&server_objects)
	for _, server := range server_objects {
		// log.Debugf("NOT updating server: %s", server.Host)
		server.Update(s)
	}

	log.Info("READY!")
}
