/* 	quote.go
_________________________________
Manages quotes for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

//QuoteAdd - takes quote and adds it to the quote array in the guild
func QuoteAdd(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	gIndex := GetGuildByID(m.GuildID) //Get guild Index

	//Add quote to quote object
	var quoteData = Quotes{data, time.Now()}

	fmt.Print(quoteData)

	//Append quote onto the guild's quote list
	kdb[gIndex].Quoted = append(kdb[gIndex].Quoted, quoteData)

	WriteKDB()
}
