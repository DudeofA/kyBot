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
	result := WaitForVotes(s, voteMsg, optionNum)

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
	guild := kdb.ReadGuild(s, m.GuildID)

	switch reaction {
	case "UPVOTE":
		err := s.MessageReactionAdd(m.ChannelID, m.ID, guild.Emotes.Upvote)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to use upvote emote, check emote config")
		}
		break

	case "DOWNVOTE":
		err := s.MessageReactionAdd(m.ChannelID, m.ID, guild.Emotes.Downvote)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to use downvote emote, check emote config")
		}
		break
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to post emote: %s", reaction))
		break
	}
}

//WaitForVotes - wait for enough votes to pass the vote or timeout and fail
func WaitForVotes(s *discordgo.Session, m *discordgo.Message, options int) (result int) {
	guild := kdb.ReadGuild(s, m.GuildID)

	voteAlive := true
	for voteAlive {
		//One options = upvote vs downvote
		if options == 0 {
			//Get each reaction
			upReact, err := s.MessageReactions(m.ChannelID, m.ID, guild.Emotes.Upvote, 10)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use upvote emote, check config")
			}
			downReact, err := s.MessageReactions(m.ChannelID, m.ID, guild.Emotes.Downvote, 10)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unable to use downvote emote, check config")
			}

			//If there are enough upvote, approve vote
			if len(upReact) > guild.Config.MinVotes {
				return 0
			}
			//If there are enough downvotes, fail vote
			if len(downReact) > guild.Config.MinVotes {
				return -1
			}

			//Otherwise 1-9 options
		} else {
			//hm
		}
		time.Sleep(1 * time.Second)
	}

	return -2
}
