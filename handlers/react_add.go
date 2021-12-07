package handlers

import (
	"kyBot/kyDB"
	"kyBot/minecraft"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func ReactAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}
	log.Debug(r.Emoji.Name)

	var server minecraft.MinecraftServer
	result := kyDB.DB.Where(&minecraft.MinecraftServer{MessageID: r.MessageID}).First(&server)
	if result.RowsAffected == 1 {
		err := s.MessageReactionsRemoveAll(r.ChannelID, r.MessageID)
		if err != nil {
			log.Errorln("Error removing emote", err.Error())
		}
		server.UpdateServer(s)
	}
}
