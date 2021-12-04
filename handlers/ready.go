package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	servers := s.State.Guilds

	logrus.Debug(self)
	for _, server := range servers {
		logrus.Debug(server.ID)
	}
}
