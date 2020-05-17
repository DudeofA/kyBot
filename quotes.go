package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

//----- Q U O T E   M A N A G E M E N T -----

// CreateQuote - create quote and insert it into the database
func (kdb *KDB) CreateQuote(s *discordgo.Session, guildID string, quoteText string) (quote Quote) {

	// Generate 3 letter identifier
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	identifier := string(b)

	// Add quote to quote object
	quote = Quote{guildID, identifier, quoteText, time.Now()}

	// Add quote to quote collection
	_, err := k.db.Exec("INSERT INTO quotes (identifier, guildID, quote, timestamp) VALUES(?,?,?,?)",
		quote.Identifier, quote.GuildID, quote.Quote, quote.Timestamp.Format("2006-01-02 15:04:05"))
	if err != nil {
		panic(err)
	}

	LogDB("Quote", quote.Identifier, quote.GuildID, "inserted into")

	return quote
}

// ReadQuote - try to get the vote from the database, returning empty quote if none found
//	returns random quote with no argument
func (kdb *KDB) ReadQuote(s *discordgo.Session, guildID, identifier string) (quote Quote) {
	var tempTime string

	if identifier == "" {
		rand := k.db.QueryRow("SELECT identifier, guildID, quote, timestamp FROM quotes ORDER BY RAND()")
		err := rand.Scan(&quote.Identifier, &quote.GuildID, &quote.Quote, &tempTime)
		switch err {
		case sql.ErrNoRows:
			LogDB("Quote", quote.Identifier, quote.GuildID, "not found in")
			return Quote{}
		case nil:
			LogDB("Quote", quote.Identifier, quote.GuildID, "read from")
			quote.Timestamp, err = time.Parse("2006-01-02 15:04:05", tempTime)
			if err != nil {
				panic(err)
			}
			return quote
		default:
			panic(err)
		}
	}

	// Search by discord guild ID & identifier
	row := k.db.QueryRow("SELECT identifier, guildID, quote, timestamp FROM quotes WHERE guildID=(?) AND identifier=(?)", guildID, identifier)
	err := row.Scan(&quote.Identifier, &quote.GuildID, &quote.Quote, &tempTime)
	switch err {
	case sql.ErrNoRows:
		LogDB("Quote", quote.Identifier, quote.GuildID, "not found")
		return Quote{}
	case nil:
		LogDB("Quote", quote.Identifier, quote.GuildID, "read from")
		quote.Timestamp, err = time.Parse("2006-01-02 15:04:05", tempTime)
		if err != nil {
			panic(err)
		}
		return quote
	default:
		panic(err)
	}
}

// UpdateID [Quote] - update identifier in quote database
func (quote *Quote) UpdateID(s *discordgo.Session, newID string) {

	k.db.Exec("UPDATE quotes SET identifier = ? WHERE identifier = ?", newID, quote.Identifier)
	quote.Identifier = newID
	LogDB("Quote", quote.Identifier, quote.GuildID, fmt.Sprintf("ID updated to %s in", newID))

}

// QuoteListIDs - List all possible quote IDs
func QuoteListIDs(s *discordgo.Session, cID string) {
	allQuotes, err := k.db.Query("SELECT identifier FROM quotes ORDER BY identifier ASC")
	if err != nil {
		panic(err)
	}

	var idenString []string
	var iden string
	for allQuotes.Next() {
		if err := allQuotes.Scan(&iden); err != nil {
			panic(err)
		}
		idenString = append(idenString, iden)
	}

	msg := strings.Join(idenString, "\n")
	s.ChannelMessageSend(cID, "```\nValid quote IDs are:\n"+msg+"\n```")
}

// RequestIdentifier - Request a custom identifier from the quote submitter
func (quote *Quote) RequestIdentifier(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Create watch entry
	vote := k.kdb.ReadVote(r.MessageID)
	k.kdb.CreateUserWatch(r.ChannelID, vote.SubmitterID, quote.Identifier)

	// Make request
	s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("<@!%s>, please type an identifier for this quote (<15 characters, no spaces)", vote.SubmitterID))
}
