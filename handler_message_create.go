package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if strings.HasPrefix(m.Content, "Wordle") {
		AddWordleStats(m.Message)
	}

	if !strings.HasPrefix(m.Content, "k!") {
		return
	}

	perm, err := s.State.MessagePermissions(m.Message)
	if err != nil {
		log.Errorf("Unable to get message author's channel: %s", err.Error())
	}
	if perm&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		log.Warnf("%s tried to use a command but their permissions are %d", m.Author.Username, perm)
		return
	}

	trim := strings.TrimPrefix(m.Content, "k!")
	split_content := strings.SplitN(trim, " ", 2)
	if len(split_content) < 1 {
		return
	}
	command := strings.ToLower(split_content[0])
	// var data string
	// if len(split_content) < 2 {
	// 	data = ""
	// } else {
	// 	data = split_content[1]
	// }

	// log.Debug(data)

	switch command {
	case "wordle":
		var wordles []Wordle
		db.Find(&wordles)

		for _, raw_wordle := range wordles {
			wordle, err := GetWordle(raw_wordle.ChannelID)
			if err != nil {
				log.Error(err)
			}
			for _, user := range wordle.Players {
				user.CalculateStats()
			}
			wordle.StatusMessageID = ""
			wordle.RefreshStatus()
		}
	}
}
