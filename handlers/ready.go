package handlers

import (
	"kyBot/commands"
	"kyBot/config"
	"kyBot/kyDB"
	"kyBot/status"
	"kyBot/update"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	guilds := s.State.Guilds

	log.Infof("%s Bot has started...running in %d Discord servers", self, len(guilds))

	// Register commands
	commands.RegisterCommands(config.APPID, s)

	kyDB.DB.AutoMigrate(&status.Server{}, &status.Wordle{}, &status.WordleStat{})

	// Loop through all Minecraft status and update their status
	var server_objects []status.Server
	_ = kyDB.DB.Not(&status.Server{Type: "wordle"}).Find(&server_objects)
	for _, server := range server_objects {
		server.Update(s)
	}

	// Migrate servers to Wordle
	update.ConvertServerToWordle(s)

	log.Info("READY!")
}
