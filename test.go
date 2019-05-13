package main

import (
	"github.com/bwmarrin/discordgo"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Testing Starting...")

	UpdateMinecraft(s, m.ChannelID)

	s.ChannelMessageSend(m.ChannelID, "Testing Done")

}
