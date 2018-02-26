package main

import (
    "github.com/bwmarrin/discordgo"
    //"fmt"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
    s.ChannelMessageSend(m.ChannelID, "Testing...")
    s.ChannelMessageSend(m.ChannelID, config.Test[1])
    s.ChannelMessageSend(m.ChannelID, config.Admin)
    s.ChannelMessageSend(m.ChannelID, config.LogID)
}
