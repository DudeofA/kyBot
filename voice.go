package main

import (
    "github.com/bwmarrin/dgvoice"
    "github.com/bwmarrin/discordgo"
)

func PlayClip(s *discordgo.Session, m *discordgo.MessageCreate, clip string) {
    c, _ := s.State.Channel(m.ChannelID)
    g, _ := s.State.Guild(c.GuildID)
    switch clip {
    case "yee":
        if !config.Yee {
            s.ChannelMessageSend(m.ChannelID, "Yee is disabled")
        } else {
            // Search through the guild's voice channels for the command's author
            for _, vs := range g.VoiceStates {
                if vs.UserID == m.Author.ID {
                    voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
                    // TO REPLACE WITH MY OWN CODE
                    dgvoice.PlayAudioFile(voiceChannel, "clips/yee.mp3", stopChan);
                    voiceChannel.Disconnect();
                }
            }
        }
        break
    case "bitconnect":
            // Search through the guild's voice channels for the command's author
            for _, vs := range g.VoiceStates {
                if vs.UserID == m.Author.ID {
                    voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
                    // TO REPLACE WITH MY OWN CODE
                    dgvoice.PlayAudioFile(voiceChannel, "clips/bitconnect.wav", stopChan);
                    voiceChannel.Disconnect();
                }
            }
        }
    }

