package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Vote(s *discordgo.Session, m *discordgo.MessageCreate, voteSubject string) (result bool) {
    //Constants
    neededUpvotes := 3
    neededDownvotes := 2
    cID := m.ChannelID
    result = false

    //Save messages to delete later
    quoteMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```css\n%s\n```", voteSubject))
    voteMsg, _ := s.ChannelMessageSend(m.ChannelID, "Vote on this quote by reacting")
    tempMsg := []string{quoteMsg.ID}

    //Add upvote and downvote reactions
    s.MessageReactionAdd(m.ChannelID, voteMsg.ID, "Upvote:402490285365657600") 
    s.MessageReactionAdd(m.ChannelID, voteMsg.ID, "Downvote:402490335789318144") 

    //Every second check votes
    ticker := time.NewTicker(time.Millisecond * 1000)
    go func() {
        for range ticker.C {
            //check for enough votes
            //Get the number of people who reacted and compare the length to the neededUp/Downvotes
            //Failed vote
            downCount, _ := s.MessageReactions(m.ChannelID, voteMsg.ID, "Downvote:402490335789318144", 10)
            if len(downCount) == neededDownvotes + 1 {
                s.ChannelMessageEdit(m.ChannelID, voteMsg.ID, "Vote failed!")
                CleanVote(s, cID, tempMsg)
                result = false
                return
            }

            //Success vote
            upCount, _ := s.MessageReactions(m.ChannelID, voteMsg.ID, "Upvote:402490285365657600", 10)
            if len(upCount) == neededUpvotes + 1 {
                s.ChannelMessageEdit(m.ChannelID, voteMsg.ID, "Vote passed!")
                CleanVote(s, cID, tempMsg)
                result = true
                return
            }
        }
    }()

    time.Sleep(time.Minute * 10)
    ticker.Stop()
    //Timed out
    //Not enough votes
    if !result {
        s.ChannelMessageEdit(m.ChannelID, voteMsg.ID, "Not enough votes")
        CleanVote(s, cID, tempMsg)
        return
    }
    return
}

func CleanVote(s *discordgo.Session, cID string, m []string) {
    s.ChannelMessagesBulkDelete(cID, m)
}


