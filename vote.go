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
	array := strings.SplitAfterN(data, " ", 1)
	options := array[0]
	text := array[1]
	optionNum, err := strconv.Atoi(options)
	if err != nil {
		panic(err)
	}

	//How many vote options
	switch optionNum {
	case 1:
		s.ChannelMessageSend(m.ChannelID, "Starting vote...Upvote/Downvote to cast your vote")
		quoteMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", text))
		s.MessageReactionAdd(m.ChannelID, quoteMsg.ID, "upvote:402490285365657600")
		s.MessageReactionAdd(m.ChannelID, quoteMsg.ID, "downvote:402490335789318144")
		//GetVoteResults(s, m, optionNum)
		break

	default:
		break
	}

	return false
}
