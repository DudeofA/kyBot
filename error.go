package main

import "github.com/bwmarrin/discordgo"

//ErrorPrint - Sends error message based on the
//type of error to the appropriate channel
func ErrorPrint(s *discordgo.Session, channelID string, msgType string) {
	errorMsg := "Unknown error type"

	switch msgType {

	case "NOPERM":
		errorMsg = "You do not have permission to use this command"
		break
	}

	s.ChannelMessageSend(channelID, errorMsg)
}
