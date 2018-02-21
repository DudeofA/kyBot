package main

import (
    "github.com/bwmarrin/discordgo"
    "github.com/bwmarrin/dgvoice"
    //"time"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
    stopChan := make(chan bool)
    c, _ := s.State.Channel(m.ChannelID)
    g, _ := s.State.Guild(c.GuildID)
    for _, vs := range g.VoiceStates {
        if vs.UserID == m.Author.ID {
            voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
            dgvoice.PlayAudioFile(voiceChannel, "clips/yee.mp3", stopChan);
            voiceChannel.Disconnect();
        }
    }
}
