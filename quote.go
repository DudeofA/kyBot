/* 	quote.go
_________________________________
Manages quotes for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

//QuoteAdd - takes quote and adds it to the quote array in the guild
func QuoteAdd(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	gIndex := GetGuildByID(m.GuildID) //Get guild Index

	//Add quote to quote object
	var quoteData = Quote{data, time.Now()}

	//Append quote onto the guild's quote list
	kdb.Servers[gIndex].Quotes = append(kdb.Servers[gIndex].Quotes, quoteData)

	//Write back to kdb
	kdb.Write()
}

//QuoteGet - Returns the quote indexed at the argument or a random one if argument if -1
func QuoteGet(m *discordgo.MessageCreate, index int) Quote {
	//Get kdb guild index for current guild
	gIndex := GetGuildByID(m.GuildID)
	//Save length of quote list (max for random generator)
	quoteLen := len(kdb.Servers[gIndex].Quotes)

	//If index is -1, return random quote
	if index == -1 && quoteLen >= 0 {
		index = rand.Intn(quoteLen)
	}

	//Return the quote as long as it is valid
	if index < quoteLen && index >= 0 {
		//Get original quote
		rawQuote := kdb.Servers[gIndex].Quotes[index]
		//Add index of quote to be displayed
		rawQuote.Quote = fmt.Sprintf("[%d]# %s", index+1, rawQuote.Quote)
		return rawQuote
	}

	//Return empty quote if fail
	return Quote{"Doesn't seem to be a quote here...", time.Time{}}
}

//QuotePrint - Prints the quote with nice colors
func QuotePrint(s *discordgo.Session, m *discordgo.MessageCreate, q Quote) *discordgo.Message {
	//Format quote with colors using the CSS formatting
	fmtQuote := fmt.Sprintf("```ini\n[ %s ]\n%s\n```", q.Timestamp.Format("Jan 2 3:04:05PM 2006"), q.Quote)
	//Return the sent message for vote monitoring
	msg, _ := s.ChannelMessageSend(m.ChannelID, fmtQuote)
	return msg
}
