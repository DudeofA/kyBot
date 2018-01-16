package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Vote(s *discordgo.Session, m *discordgo.MessageCreate, voteSubject string) (result bool) {
    //Constants
    neededUpvotes := 2
    neededDownvotes := 1
    cID := m.ChannelID

    //Save messages to delete later
    quoteMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```css\n%s\n```", voteSubject))
    voteMsg, _ := s.ChannelMessageSend(m.ChannelID, "Vote on this quote by reacting (10 seconds)")
    tempMsg := []string{quoteMsg.ID, voteMsg.ID}

    //Add upvote and downvote reactions
    s.MessageReactionAdd(m.ChannelID, voteMsg.ID, "Upvote:402490285365657600") 
    s.MessageReactionAdd(m.ChannelID, voteMsg.ID, "Downvote:402490335789318144") 
    //wait 10 seconds - NEEDS WORKS
    time.Sleep(10 * time.Second)

    //Get the number of people who reacted and compare the length to the neededUp/Downvotes
    //Failed vote
    downCount, _ := s.MessageReactions(m.ChannelID, voteMsg.ID, "Downvote:402490335789318144", 10)
    if len(downCount) == neededDownvotes + 1 {
        s.ChannelMessageSend(m.ChannelID, "Vote failed!")
        CleanVote(s, cID, tempMsg)
        result = false
        return
    }

    //Success vote
    upCount, _ := s.MessageReactions(m.ChannelID, voteMsg.ID, "Upvote:402490285365657600", 10)
    if len(upCount) == neededUpvotes + 1 {
        s.ChannelMessageSend(m.ChannelID, "Vote passed!")
        CleanVote(s, cID, tempMsg)
        result = true
        return
    }

    //Not enough votes
    s.ChannelMessageSend(m.ChannelID, "Not enough votes")
    CleanVote(s, cID, tempMsg)
    result = false
    return
}

func CleanVote(s *discordgo.Session, cID string, m []string) {
    s.ChannelMessagesBulkDelete(cID, m)
}


