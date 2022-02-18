package handlers

import (
	"fmt"
	"kyBot/component"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if strings.HasPrefix(m.Content, "Wordle") {
		component.AddWordleStats(s, m.Message, "")
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
	var data string
	if len(split_content) < 2 {
		data = ""
	} else {
		data = split_content[1]
	}

	switch command {
	case "import":
		regex := regexp.MustCompile(`^https://discord\\.com/channels/(.*)/(.*)/(.*)`)

		messageLink := regex.FindStringSubmatch(data)
		if len(messageLink) != 4 {
			s.ChannelMessageSend(m.ChannelID, "Not a valid discord message link")
			return
		}
		// messageLink[0] is the whole link
		// messageLink[1] is the guild ID
		chanID := messageLink[2]
		msgID := messageLink[3]

		if err := component.ImportWordleStat(s, m.ChannelID, chanID, msgID); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to add Wordle stat: %s", err))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Added Wordle stats successfully")
		}
	case "scrape":
		component.ScrapeChannel(s, m.ChannelID)
		log.Debug("Done scraping")
	case "wordle":
		component.SendWordleReminders(s)
	}
}
