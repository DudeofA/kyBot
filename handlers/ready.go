package handlers

import (
	"kyBot/minecraft"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	self := event.User
	servers := s.State.Guilds

	log.Infof("%s Bot has started...running in %d Discord servers", self, len(servers))

	minecraft.UpdateAllServers(s)
}
