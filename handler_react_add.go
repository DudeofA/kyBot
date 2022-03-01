package main

import (
	"github.com/bwmarrin/discordgo"
)

func ReactAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}
}
