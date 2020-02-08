package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
