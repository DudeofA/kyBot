/* 	vote.go
_________________________________
Parses votes and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

//startVote - begin a vote with variable vote options
func startVote(s *discordgo.Session, m *discordgo.MessageCreate, data string) int {
	//Parse the incoming command into # of vote options and string afterward
	array := strings.SplitAfter(data, " ")
	options := array[0]
	optionNum, err := strconv.Atoi(strings.TrimSpace(options))
	if err != nil {
		panic(err)
	}
	//Take off formatting
	text := strings.TrimLeft(data, options)

	//Declare variable so that it persists throughout the switch options
	var voteMsg *discordgo.Message

	//How many vote options
	switch optionNum {
	case 0:
		s.ChannelMessageSend(m.ChannelID, "Starting vote...cast your vote now!")
		//Send and save the vote message to be modified later
		quote := Quote{m.GuildID, "identifier", text, time.Now()}
		voteMsg = QuotePrint(s, m, quote)
		ReactionAdd(s, voteMsg, "UPVOTE")
		ReactionAdd(s, voteMsg, "DOWNVOTE")
		break

	default:
		break
	}

	//Remove original vote post
	s.ChannelMessageDelete(m.ChannelID, m.ID)

	//Pend on the votes until they pass or timeout
	result := WaitForVotes(s, voteMsg, m.Author, optionNum)

	switch result {
	case -1:
		s.ChannelMessageSend(m.ChannelID, "Vote failed, yikes")
		break

	case 0:
		s.ChannelMessageSend(m.ChannelID, "Vote succeeded, yay!")
		break

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Option %d wins the vote!", result))
	}

	return result
}

//ReactionAdd - add a reaction to the passed-in message
func ReactionAdd(s *discordgo.Session, m *discordgo.Message, reaction string) {

	switch reaction {
	case "UPVOTE":
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "⬆️")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to use upvote emote")
		}
		break

	case "DOWNVOTE":
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "⬇️")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to use downvote emote")
		}
		break
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to post emote: %s", reaction))
		break
	}
}

//WaitForVotes - wait for enough votes to pass the vote or timeout and fail
func WaitForVotes(s *discordgo.Session, m *discordgo.Message, author *discordgo.User, options int) (result int) {

	voteAlive := true
	for voteAlive {
		//One options = upvote vs downvote
		if options == 0 {
			//Get each reaction
			upReact, err := s.MessageReactions(m.ChannelID, m.ID, "⬆️", 10)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to check upvote emote")
			}
			downReact, err := s.MessageReactions(m.ChannelID, m.ID, "⬇️", 10)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to check downvote emote")
			}

			//If there are enough upvote, approve vote
			if len(upReact) > k.botConfig.MinVotes {
				return 0
			}
			//If there are enough downvotes, fail vote
			if len(downReact) > k.botConfig.MinVotes-1 {
				return -1
			}
			// If original user downvotes, fail vote
			for _, a := range downReact {
				if a.ID == author.ID {
					return -1
				}
			}

			//Otherwise 1-9 options
		} else {
			//hm
		}
		time.Sleep(1 * time.Second)
	}

	return -2
}
