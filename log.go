package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetAuthor(s *discordgo.Session, i interface{}, code string) (name string) {
	switch code {
	// For each type of event, gets the user and member structs
	case "MSG":
		// Typecast interface to MessageCreate
		m := i.(*discordgo.MessageCreate)
		//Get user and member using what is within each event struct
		user := m.Author
		channel, _ := s.State.Channel(m.ChannelID)
		member, err := s.State.Member(channel.GuildID, m.Author.ID)
		name = FormatAuthor(user, member, err)
		break
	case "STATUS":
		// Typecast interface to PresenceUpdate
		p := i.(*discordgo.PresenceUpdate)
		//Get user and member using what is within each event struct
		user := p.User
		member, err := s.State.Member(p.GuildID, user.ID)
		name = FormatAuthor(user, member, err)
		break
	case "VOICE":
		// Typecast interface to VoiceStateUpdate
		v := i.(*discordgo.VoiceStateUpdate)
		//Get user and member using what is within each event struct
		user, _ := s.User(v.UserID)
		//Get channel despite if the voicestateupdate is empty
		var channel *discordgo.Channel
		if v.ChannelID == "" {
			fileUser, _ := ReadUser(s, v, "VOICE")
			channel, _ = s.State.Channel(fileUser.LastSeenCID)
		} else {
			channel, _ = s.State.Channel(v.ChannelID)
		}
		member, err := s.State.Member(channel.GuildID, v.UserID)
		name = FormatAuthor(user, member, err)
		break
	default:
		s.ChannelMessageSend(config.LogID, "GetAuthor failed")
	}

	return name
}

func FormatAuthor(user *discordgo.User, member *discordgo.Member, err error) (name string) {
	// Gets nickname with full username in parentheses
	if err != nil {
		name = user.Username + "#" + user.Discriminator
	} else {
		if member.Nick != "" {
			name = member.Nick + " " + "(" + user.Username + "#" + user.Discriminator + ")"
		} else {
			name = user.Username + "#" + user.Discriminator
		}
	}
	return name
}

func Log(s *discordgo.Session, i interface{}, code string) {
	if config.LogID == "" {
		return
	}
	timestamp := time.Now()
	timestampf := timestamp.Format("Mon Jan 2 - 3:04PM")

	switch code {

	case "MSG":
		if !config.LogMessage {
			return
		}
		m := i.(*discordgo.MessageCreate)
		username := GetAuthor(s, i, code)

		channel, _ := s.Channel(m.ChannelID)
		s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s - %s:\n!MSG: %s\n```",
			timestampf, channel.Name, username, code, m.ContentWithMentionsReplaced()))
		break

	case "STATUS":
		if !config.LogStatus {
			return
		}
		p := i.(*discordgo.PresenceUpdate)
		username := GetAuthor(s, i, code)
		s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s:\n!STATUS: %s\n```",
			timestampf, username, code, p.Status))
		break

	case "VOICE":
		if !config.LogVoice {
			return
		}
		v := i.(*discordgo.VoiceStateUpdate)
		username := GetAuthor(s, i, code)
		var action string
		var channel *discordgo.Channel
		if v.ChannelID == "" {
			action = "Left"
			fileUser, _ := ReadUser(s, v, "VOICE")
			channel, _ = s.State.Channel(fileUser.LastSeenCID)
		} else {
			action = "Joined"
			channel, _ = s.State.Channel(v.ChannelID)
		}
		s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s:\n!VOICE: %s: %s\n```",
			timestampf, username, code, action, channel.Name))
	default:
		break
	}
}
