package handlers

import (
	"kyBot/status"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if !strings.HasPrefix(m.Content, "k!") {
		return
	}

	trim := strings.TrimPrefix(m.Content, "k!")
	split_content := strings.SplitN(trim, " ", 2)
	if len(split_content) < 2 {
		return
	}
	command := strings.ToLower(split_content[0])
	// data := split_content[1]

	switch command {
	case "test":
		status.SendWordleReminders(s)
	}
}
