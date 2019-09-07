/* 	quote.go
_________________________________
Manages quotes for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
