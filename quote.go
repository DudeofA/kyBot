/* 	quote.go
_________________________________
Manages quotes for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

// QuoteAdd - takes quote and adds it to the quote array in the guild
func QuoteAdd(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	guild := kdb.ReadGuild(s, m.GuildID) //Get guild Index

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

	// Generate 3 letter identifier
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	identifier := string(b)

	// Add quote to quote object
	var quoteData = Quote{m.GuildID, identifier, data, time.Now()}

	// Append quote onto the guild's quote list
	kdb.InsertQuote(quoteData)
}

// QuoteGet - Returns the quote indexed at the argument or a random one if argument if -1
func QuoteGet(s *discordgo.Session, m *discordgo.MessageCreate, identifier string) Quote {
	// Get kdb guild index for current guild
	guild := kdb.ReadGuild(s, m.GuildID)
	// Save length of quote list (max for random generator)
	quoteLen64, err := kdb.QuoteColl.EstimatedDocumentCount(context.Background())
	if err != nil {
		panic(err)
	}
	quoteLen := int(quoteLen64)

	// If idenifier is blank, return a random quote
	return kdb.ReadQuote("")
}

// QuotePrint - Prints the quote with nice colors
func QuotePrint(s *discordgo.Session, m *discordgo.MessageCreate, q Quote) *discordgo.Message {
	// Format quote with colors using the CSS formatting
	fmtQuote := fmt.Sprintf("```ini\n[ %s ]\n%s\n```", q.Timestamp.Format("Jan 2 3:04:05PM 2006"), q.Quote)
	// Return the sent message for vote monitoring
	msg, _ := s.ChannelMessageSend(m.ChannelID, fmtQuote)
	return msg
}
