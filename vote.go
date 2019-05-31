/* 	vote.go
_________________________________
Parses commands and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//startVote - begin a vote with variable vote options
func startVote(s *discordgo.Session, m *discordgo.MessageCreate, data string) bool {
	array := strings.SplitAfter(data, " ")
	options := array[0]
	text := array[1]

	switch options {
	case "1":
		s.ChannelMessageSend(m.ChannelID, "Starting vote...Upvote/Downvote to cast your vote")
		quoteMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", text))
		s.MessageReactionAdd(m.ChannelID, quoteMsg.ID, "upvote:402490285365657600")
		s.MessageReactionAdd(m.ChannelID, quoteMsg.ID, "downvote:402490335789318144")
		break

	default:
		break
	}

	return false
}
