/* 	vote.go
_________________________________
Parses commands and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//startVote - begin a vote with variable vote options
func startVote(s *discordgo.Session, m *discordgo.MessageCreate, data string) bool {
	//Parse the incoming command into data and strings
	array := strings.SplitAfter(data, " ")
	options := array[0]
	optionNum, err := strconv.Atoi(strings.TrimSpace(options))
	if err != nil {
		panic(err)
	}
	//Take off formatting
	text := strings.TrimLeft(data, options)

	//How many vote options
	switch optionNum {
	case 1:
		s.ChannelMessageSend(m.ChannelID, "Starting vote...Upvote/Downvote to cast your vote")
		voteMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", text))
		ReactionAdd(s, voteMsg, "UPVOTE")
		ReactionAdd(s, voteMsg, "DOWNVOTE")

		break

	default:
		break
	}

	return false
}

//ReactionAdd - add a reaction to the passed-in message
func ReactionAdd(s *discordgo.Session, m *discordgo.Message, reaction string) {
	switch reaction {
	case "UPVOTE":
		if kdb[GetGuildByID(m.GuildID)].Emotes.UPVOTE != "" {
			err := s.MessageReactionAdd(m.ChannelID, m.ID, kdb[GetGuildByID(m.GuildID)].Emotes.UPVOTE)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use upvote emote, check custom emotes")
			}
		} else {
			err := s.MessageReactionAdd(m.ChannelID, m.ID, "⬆")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use fallback upvote emote, that is bad")
			}
		}
		break

	case "DOWNVOTE":
		if kdb[GetGuildByID(m.GuildID)].Emotes.DOWNVOTE != "" {
			err := s.MessageReactionAdd(m.ChannelID, m.ID, kdb[GetGuildByID(m.GuildID)].Emotes.DOWNVOTE)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use downvote emote, check custom emotes")
			}
		} else {
			err := s.MessageReactionAdd(m.ChannelID, m.ID, "⬇")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use fallback downvote emote, that is bad")
			}
		}
		break
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to post emote: %s", reaction))
		break
	}
}
