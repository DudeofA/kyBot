package main

import (
	"github.com/bwmarrin/discordgo"
)

func Quote(s *discordgo.Session, message *discordgo.Message) {
	s.ChannelMessageSend(message.ChannelID, message.ContentWithMentionsReplaced())
}
