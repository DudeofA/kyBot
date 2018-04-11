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

	ResetDailies()

	//Makes sure it makes it to then end
	//
	s.ChannelMessageSend(m.ChannelID, "Testing Complete.")
}
