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

	"github.com/bwmarrin/discordgo"
)

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
	fmtQuote := fmt.Sprintf("```ini\n[ %s ] - [ %s ]\n%s\n```", q.Timestamp.Format("Jan 2 3:04:05PM 2006"), q.Identifier, q.Quote)
	// Return the sent message for vote monitoring
	msg, err := s.ChannelMessageSend(m.ChannelID, fmtQuote)
	if err != nil {
		panic(err)
	}
	return msg
}
