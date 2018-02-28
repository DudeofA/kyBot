package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Log(s *discordgo.Session, i interface{}, code string) {

        namestr := ""
        timestamp := time.Now()
        timestampf := timestamp.Format("Mon Jan 2 - 3:04PM")

	switch code {

	case "MSG":
        m := i.(*discordgo.MessageCreate)
        channel, _ := s.State.Channel(m.ChannelID)
        member, err := s.State.Member(channel.GuildID, m.Author.ID)
        // Ugly code to make the formatting pretty
        if err != nil {
            namestr = m.Author.Username + "#" + m.Author.Discriminator
        } else {
            if member.Nick != "" {
                namestr = member.Nick + " " + "(" + m.Author.Username + "#" + m.Author.Discriminator + ")"
            } else {
                namestr = m.Author.Username + "#" + m.Author.Discriminator
            }
        }

        channelType, _ := s.Channel(m.ChannelID)
        channelName := channelType.Name
		s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s - %s:\n!MSG: %s\n```", timestampf, channelName, namestr, code, m.Content))
		break

    case "STATUS":
        p := i.(*discordgo.PresenceUpdate)

        user , _ := s.State.Member(p.GuildID, p.User.ID)
        name := "No name found"
        if user.Nick != "" {
            name = user.Nick
        } else {
            name = p.User.Username
        }

        s.ChannelMessageSend(config.LogID, fmt.Sprintf("```diff\n- %s - %s - %s:\n!STATUS: %s\n```", timestampf, name, code, p.Status))


	default:
		break
	}
}
