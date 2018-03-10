package main

import (
    "github.com/bwmarrin/dgvoice"
    "github.com/bwmarrin/discordgo"
)

func PlayClip(s *discordgo.Session, m *discordgo.MessageCreate, clip string) {
    if !config.Noise {
        s.ChannelMessageSend(m.ChannelID, "Noise commands are disabled")
        return
    }

    c, _ := s.State.Channel(m.ChannelID)
    g, _ := s.State.Guild(c.GuildID)
    switch clip {
    case "yee":
        // Search through the guild's voice channels for the command's author
        for _, vs := range g.VoiceStates {
            if vs.UserID == m.Author.ID {
                voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
                // TO REPLACE WITH MY OWN CODE
                dgvoice.PlayAudioFile(voiceChannel, "clips/yee.mp3", stopChan);
                voiceChannel.Disconnect();
                return
            }
        }
        s.ChannelMessageSend(m.ChannelID, "I can't sing for you if you aren't in a voice channel :c")
        break

    case "bitconnect":
        // Search through the guild's voice channels for the command's author
        for _, vs := range g.VoiceStates {
            if vs.UserID == m.Author.ID {
                voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
                // TO REPLACE WITH MY OWN CODE
                dgvoice.PlayAudioFile(voiceChannel, "clips/bitconnect.wav", stopChan);
                voiceChannel.Disconnect();
                return
            }
        }
        s.ChannelMessageSend(m.ChannelID, "I can't sing for you if you aren't in a voice channel :c")
        break
    }
}

