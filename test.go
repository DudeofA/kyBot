package main

import (
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "os"

	"github.com/bwmarrin/discordgo"
	// "github.com/jasonlvhit/gocron"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Starting testing...")
	//
	//Make sure it actually starts and runs once

	g, _ := s.Guild(USArray.GID)
	voiceChan := config.DefaultChan
	for i := range g.VoiceStates {
		if g.VoiceStates[i].UserID == m.Author.ID {
			voiceChan = g.VoiceStates[i].ChannelID
		}
	}
	curChan, _ = s.ChannelVoiceJoin(USArray.GID, voiceChan, false, false)
	// curChan.Disconnect()

	//Makes sure it makes it to then end
	//
	s.ChannelMessageSend(m.ChannelID, "Testing Complete.")
}
