package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func logMessage(s *discordgo.Session, timestamp time.Time, user *discordgo.User, mID, cID, code, message string) {

	timestampf := timestamp.Format("Mon Jan 2 - 3:04PM")

	var namestr string
	channel, _ := s.State.Channel(cID)
	member, err := s.State.Member(channel.GuildID, user.ID)
	if err != nil {
		namestr = user.Username + "#" + user.Discriminator
	} else {
		if member.Nick != "" {
			namestr = member.Nick + " " + "(" + user.Username + "#" + user.Discriminator + ")"
		} else {
			namestr = user.Username + "#" + user.Discriminator
		}
	}

	switch code {
	case "MSG":
        channelType, _ := s.Channel(cID)
        channelName := channelType.Name
		s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s - %s:\n!MSG: %s\n```", timestampf, channelName, namestr, code, message))
		break

//	case "EDI":
//		oldMsg, _ := s.State.Message(cID, mID)
//		s.ChannelMessageSend(logID, fmt.Sprintf("```diff\n- %s - %s - %s:\n!OLD: %s\n!NEW: %s\n```", timestampf, namestr, code, oldMsg.ContentWithMentionsReplaced(), message))
//		break

//	case "DEL":
//		oldMsg, _ := s.State.Message(cID, mID)
//		s.ChannelMessageSend(logID, fmt.Sprintf("```diff\n- %s - %s - %s:\n!MSG: %s\n```", timestampf, namestr, code, oldMsg))
//		break
	default:
		break
	}
}
