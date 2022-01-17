package status

import (
	"kyBot/commands"
	"kyBot/kyDB"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	STOP_EMOJI   = "🛑"
	WORDLE_EMOJI = "🟩"
	WORDLE_URL   = "https://www.powerlanguage.co.uk/wordle/"
)

func init() {
	addServerCommand := &discordgo.ApplicationCommand{
		Name:        "add-wordle-reminder",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Add a repeating wordle reminder message",
	}
	commands.AddCommand(addServerCommand)
}

func AddWordleReminder(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "A reminder message will be posted every day in this channel",
		},
	}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		log.Errorf("Error responding to the interaction: %s", err.Error())
	}

	server := Server{
		Type:            "wordle",
		Host:            i.ChannelID,
		Port:            0,
		StatusChannelID: i.ChannelID,
	}

	if err := kyDB.DB.Take(&server).Error; err != nil {
		kyDB.DB.Create(&server)
	}
}

func SendWordleReminders(s *discordgo.Session) {
	wordleEmbed := &discordgo.MessageEmbed{
		URL:         WORDLE_URL,
		Title:       "WORDLE REMINDER",
		Description: "New wordle now available!",
		Color:       0,
	}

	stopButton := &discordgo.Button{
		Label: "Stop reminding",
		Style: 4,
		Emoji: discordgo.ComponentEmoji{
			Name:     STOP_EMOJI,
			ID:       "",
			Animated: false,
		},
		CustomID: "delete_server",
	}

	wordleLinkButton := &discordgo.Button{
		Label: "Wordle Link",
		Style: 5,
		URL:   WORDLE_URL,
		Emoji: discordgo.ComponentEmoji{
			Name:     WORDLE_EMOJI,
			ID:       "",
			Animated: false,
		},
	}

	msg := &discordgo.MessageSend{
		Embed: wordleEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					wordleLinkButton,
					stopButton,
				},
			},
		},
	}

	var server_objects []Server
	_ = kyDB.DB.Where(&Server{Type: "wordle"}).Find(&server_objects)
	for _, server := range server_objects {
		_, err := s.ChannelMessageSendComplex(server.StatusChannelID, msg)
		if err != nil {
			log.Errorf("Error sending wordle update: %s", err.Error())
		}
	}
}
